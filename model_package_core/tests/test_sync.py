import json

from model_package_core.constants import METADATA_FILENAME, WEIGHT_FILENAME
from model_package_core.sync import ModelMetadataSync
from model_package_core.tests.conftest import create_peer_dir


def test_new_db_syncs_all_peers(peers_db, peer_models_path):
    # GIVEN an empty peers db and a peers models path that has 3 peers
    assert len(peers_db) == 0

    # WHEN we generate the models to train
    sync_commander = ModelMetadataSync(db=peers_db, peer_models=peer_models_path)
    models_to_train = sync_commander._models_to_train()

    # THEN
    peers_in_peer_path = []
    for peer_path in peer_models_path.iterdir():
        peer_id = peer_path.name
        peers_in_peer_path.append(peer_id)

        # the DB should be synced with the latest values
        assert peers_db[peer_id] is not None
        with open(peer_path / METADATA_FILENAME, 'r') as f:
            assert peers_db[peer_id] == json.load(f)

        # AND the models to train should have the correct data
        assert peers_db[peer_id]['sample_size'] == models_to_train[peer_id]['number_of_samples']
        assert models_to_train[peer_id]['weights'] == peer_path / WEIGHT_FILENAME

    # AND the models to train dict and DB has all the new peers
    assert len(models_to_train) == len(peers_in_peer_path) == len(peers_db)


def test_syncs_new_peer_added_to_db(db_with_peers, peer_models_path):
    # GIVEN a db with three peers
    assert len(db_with_peers) == 3

    # WHEN a new peer is added
    new_peer_id = 'test_new_peer'
    metadata = {"version": 1, "sample_size": 10, "last_updated": 1657843681}
    create_peer_dir(peer_id=new_peer_id, peer_metadata=metadata, peers_path=peer_models_path)
    # AND we sync the db
    sync_commander = ModelMetadataSync(db=db_with_peers, peer_models=peer_models_path)
    models_to_train = sync_commander._models_to_train()

    # THEN models to train references the new peer only
    assert len(models_to_train) == 1
    assert new_peer_id in models_to_train
    assert models_to_train[new_peer_id]['number_of_samples'] == 10

    # AND the db also references the new peer
    assert len(db_with_peers) == 4
    assert db_with_peers[new_peer_id] == metadata


def test_syncs_existing_peer_new_model_version_updates_db(db_with_peers, peer_models_path):
    # GIVEN a sync to db with three peers
    assert len(db_with_peers) == 3

    # WHEN a peer is updated with a new model version
    updated_peer_id = 'peer-2'
    assert db_with_peers[updated_peer_id]["version"] == 2

    # update metadata
    metadata_file = peer_models_path / updated_peer_id / METADATA_FILENAME
    with open(metadata_file, "r") as json_file:
        metadata = json.load(json_file)

    metadata["version"] = 3

    with open(metadata_file, "w") as json_file:
        json.dump(metadata, json_file)

    # AND we sync the db
    sync_commander = ModelMetadataSync(db=db_with_peers, peer_models=peer_models_path)
    models_to_train = sync_commander._models_to_train()

    # THEN models to train references the updated peer only
    assert len(models_to_train) == 1
    assert updated_peer_id in models_to_train

    # AND the db also references the updated peer
    assert len(db_with_peers) == 3
    assert db_with_peers[updated_peer_id] == metadata
