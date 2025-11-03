import asyncio
from typing import Any

import discord
import duckdb
from discord import app_commands
from pydantic_ai import Agent, ModelRetry
from pydantic_ai.models.openai import OpenAIChatModel
from pydantic_ai.providers.openai import OpenAIProvider
from pydantic_settings import BaseSettings, SettingsConfigDict


class Config(BaseSettings):
    model_config = SettingsConfigDict(
        env_prefix="DUCKDBOT_STATS_",
        env_file=".env",
    )

    db_path: str = "../activity.duckdb"
    discord_token: str
    guild_id: int


config = Config()  # type: ignore
db = duckdb.connect(config.db_path, read_only=True)

db.sql("SET enable_external_access = false")
db.sql("SET allow_community_extensions = false")
db.sql("SET lock_configuration = true")

schema = "\n".join(
    [table[0] for table in db.sql("SELECT sql FROM duckdb_tables()").fetchall()]
)

model = OpenAIChatModel(
    "Qwen3-Coder-30B",
    provider=OpenAIProvider(base_url="http://llama.internal.bootleg.technology/v1"),
)

system_prompt = f"""You are a helpful assistant that answers questions about Discord server activity stored in a DuckDB database.

The database has the following schema:
```sql
{schema}
```

The data is not real-time. The last updated timestamp is stored in the `meta` table.

Your response will be returned as a Discord message - format your message using Discord's markdown where appropriate.
"""

stats_agent = Agent(
    model,
    output_type=str,
    system_prompt=system_prompt,
)


@stats_agent.tool_plain(retries=2)
async def query_db(sql: str) -> list[tuple[Any, ...]]:
    """Executes a SQL query against the DuckDB database and returns the results."""
    try:
        query_resp = await asyncio.to_thread(db.sql, sql)
        return await asyncio.to_thread(query_resp.fetchall)
    except duckdb.DatabaseError as e:
        raise ModelRetry(f"An error occurred making the provided query: {e}") from e


guild = discord.Object(id=config.guild_id)


class StatsBotClient(discord.Client):
    def __init__(self, *, intents: discord.Intents):
        super().__init__(intents=intents)
        self.tree = app_commands.CommandTree(self)

    async def setup_hook(self):
        await self.tree.sync(guild=guild)


intents = discord.Intents.default()
client = StatsBotClient(intents=intents)


@client.tree.command(guild=guild)
async def query(interaction: discord.Interaction, question: str):
    """Ask a question about Discord server activity."""
    await interaction.response.defer()
    result = await stats_agent.run(question)
    await interaction.followup.send(result.output)


client.run(config.discord_token)
