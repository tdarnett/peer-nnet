"""
This script is meant to be run as a cron task.
"""

from sqlitedict import SqliteDict
from model_package_core.sync import ModelMetadataSync

# initialize DB
database = SqliteDict("./peer_metadata.sqlite")

# run the sync
sync_commander = ModelMetadataSync(db=database)
sync_commander.run()
