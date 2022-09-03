import json

import pytest
from sqlitedict import SqliteDict

from model_package_core.constants import METADATA_FILENAME, WEIGHT_FILENAME


@pytest.fixture()
def db(tmp_path_factory):
    db_path = tmp_path_factory.mktemp('db') / 'test_peer_sync.sqlite'
    db = SqliteDict(db_path)
    yield db
    db.close()


@pytest.fixture(scope="function")
def peers_db(db):
    """
    An empty sqlitedict db object
    """
    db.clear()
    return db


@pytest.fixture(scope="session")
def peer_models_path(tmp_path_factory):
    """List of various sample peers"""
    peer_models_path = tmp_path_factory.mktemp('peer_models')

    peers = [
        {"version": 1, "sample_size": 1000, "last_updated": 1657849384},
        {"version": 2, "sample_size": 150, "last_updated": 1657849323},
        {"version": 4, "sample_size": 1300, "last_updated": 1657844553},
    ]

    # construct the fs layout
    for idx, peer_metadata in enumerate(peers):
        peer_path = peer_models_path / f'peer-{idx}'
        peer_path.mkdir()

        # save metadata as json file
        metadata_file = peer_path / METADATA_FILENAME
        with open(metadata_file, 'w') as outfile:
            outfile.write(json.dumps(peer_metadata))

        # create empty weights file
        weights_file = peer_path / WEIGHT_FILENAME
        weights_file.touch()

    return peer_models_path
