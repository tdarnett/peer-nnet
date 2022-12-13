from model_manager.constants import (TRAIN_IMAGE_DATA_PATH,
                                          TRAIN_LABEL_DATA_PATH)
from model_manager.pytorch_model.dataset import ProcessedDataset


def test_dataset_normalizes_training_data_samples():
    # GIVEN a path to raw training data images and labels
    # and a specified number of training samples
    from pathlib import PosixPath
    assert type(TRAIN_IMAGE_DATA_PATH) == PosixPath
    assert type(TRAIN_LABEL_DATA_PATH) == PosixPath
    number_of_samples = 320

    # WHEN a new ProcessedDataset is created
    dataset = ProcessedDataset(
        TRAIN_IMAGE_DATA_PATH,
        TRAIN_LABEL_DATA_PATH,
        number_of_samples
    )

    # THEN 320 samples have been loaded
    assert len(dataset) == number_of_samples

    # AND each sample has been normalized to 0 - 1
    # NOTE - this assertion is specific to the problem at hand i.e. image classification
    sample, _ = dataset[0]
    assert all(idx >= 0. for idx in sample)
    assert all(idx <= 1. for idx in sample)
