import json
from pathlib import Path

import pytest
from model_manager.constants import METADATA_FILENAME, WEIGHT_FILENAME
from model_manager.pytorch_model.early_stopping import EarlyStopping
from model_manager.pytorch_model.model import Net
from model_manager.sync import ModelMetadataSync
from sqlitedict import SqliteDict


@pytest.fixture()
def train_label_path():
    return Path('model_manager/tests/fixtures/train_labels.npy')


@pytest.fixture()
def train_image_path():
    return Path('model_manager/tests/fixtures/train_images.npy')


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


@pytest.fixture(scope="function")
def peer_models_path(tmp_path_factory):
    """List of various sample peers"""
    peer_models_path = tmp_path_factory.mktemp('peer_models')

    peers = {
        'peer-1': {"version": 1, "sample_size": 1000, "last_updated": 1657849384},
        'peer-2': {"version": 2, "sample_size": 150, "last_updated": 1657849323},
        'peer-3': {"version": 4, "sample_size": 1300, "last_updated": 1657844553},
    }

    # construct the fs layout
    for peer_id, peer_metadata in peers.items():
        create_peer_dir(peer_id=peer_id, peers_path=peer_models_path,
                        peer_metadata=peer_metadata)

    return peer_models_path


@pytest.fixture(scope='function')
def db_with_peers(peers_db, peer_models_path):
    sync_commander = ModelMetadataSync(
        db=peers_db, peer_models=peer_models_path)
    # calculate models to train on peers path to populate db
    sync_commander._models_to_train()

    return peers_db


@pytest.fixture()
def local_model_path(tmp_path_factory):
    # create a temporary file path for storing local model training files
    local_model_path = tmp_path_factory.mktemp('model')
    return local_model_path


@pytest.fixture(scope='session')
def create_early_stopping():
    # instantiate a dummy early stopping object
    early_stopping = EarlyStopping(tolerance=3, min_delta=1)
    return early_stopping


@pytest.fixture(scope='session')
def create_model():
    # instantiate a dummy model
    model = Net(
        input_size=784,
        hidden_units=[32, 16],
        number_of_classes=10
    )
    return model


def create_peer_dir(peer_id: str, peers_path, peer_metadata: dict):
    peer_path = peers_path / peer_id
    peer_path.mkdir()

    # save metadata as json file
    metadata_file = peer_path / METADATA_FILENAME
    with open(metadata_file, 'w') as outfile:
        outfile.write(json.dumps(peer_metadata))

    # create empty weights file
    weights_file = peer_path / WEIGHT_FILENAME
    weights_file.touch()
