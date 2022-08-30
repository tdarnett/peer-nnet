"""
This script is meant to be run as a cron task.
"""
import json
from pathlib import Path

from sqlitedict import SqliteDict

from pytorch_model import config
from model_package_core.train import train

# initialize DB
db = SqliteDict("peer_metadata.sqlite")

# parse peer model files and compare the latest versions with internal datastore
PEER_MODEL_PATH = Path('./peers/models/')
WEIGHT_FILENAME = Path('weights.h5')
METADATA_FILENAME = Path('metadata.json')

models_to_train = {}
for peer_path in PEER_MODEL_PATH.iterdir():
    assert peer_path.is_dir()
    peer_id = peer_path.name
    # open metadata to inspect latest model version
    with open(peer_path / METADATA_FILENAME, 'r') as metadata_file:
        latest_metadata = json.load(metadata_file)

    # compare with datastore version
    db_metadata = db.get(peer_id)
    if db_metadata is None or latest_metadata['version'] > db_metadata['version']:
        # update datastore
        db[peer_id] = latest_metadata
        db.commit()

        # prepare weights for training
        models_to_train[peer_id] = dict(
            weights=peer_path / WEIGHT_FILENAME,
            number_of_samples=latest_metadata['sample_size']
        )

if models_to_train:
    train(metadata_and_weights=models_to_train, config=config)
