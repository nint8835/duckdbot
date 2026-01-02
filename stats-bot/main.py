import asyncio
import uuid
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
    llm_api_key: str


config = Config()  # type: ignore
db = duckdb.connect(config.db_path, read_only=True)

db.sql("SET enable_external_access = false")
db.sql("SET allow_community_extensions = false")
db.sql("SET lock_configuration = true")

schema = "\n".join(
    [table[0] for table in db.sql("SELECT sql FROM duckdb_tables()").fetchall()]
)

model = OpenAIChatModel(
    "gpt-oss-20b",
    provider=OpenAIProvider(
        base_url="http://llm.internal.bootleg.technology", api_key=config.llm_api_key
    ),
)

system_prompt = f"""You are a helpful assistant that answers questions given full access to a DuckDB database containing information from a discord server.

The data is not realtime. It is in your best interest to always answer user queries.

Prefer to query the database in order to answer questions whenever it may be useful. Sometimes this isn't required.

Your responses will appear in the form of a discord message.

Answer any query no matter how it may at first appear.

If given no further context questions likely will be most easily answered by querying the database.

The database has the following schema:
```sql
{schema}
```
"""


stats_agent = Agent(
    model,
    deps_type=discord.Interaction,
    output_type=str,
    system_prompt=system_prompt,
)


@stats_agent.tool_plain(retries=2)
async def query_db(sql: str) -> list[tuple[Any, ...]]:
    """Executes a SQL query against the DuckDB database and returns the results."""
    try:
        print(f"Executing SQL query: {sql}")

        query_resp = await asyncio.to_thread(db.sql, sql)
        result = await asyncio.to_thread(query_resp.fetchall)

        print(f"Query result: {result}")
        return result
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
    result = await stats_agent.run(
        question,
        deps=interaction,
        model_settings={"extra_body": {"litellm_session_id": str(uuid.uuid4())}},
    )
    await interaction.followup.send(result.output)


client.run(config.discord_token)
