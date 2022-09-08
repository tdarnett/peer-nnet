# import the necessary packages
from torch.utils.data import Dataset
from mlxtend.data import loadlocal_mnist


class ProcessedDataset(Dataset):
    """Processing steps for dataset."""

    def __init__(self, images_path, labels_path, number_of_samples):
        """Store the image and label filepaths.

        :param images_path: Path to the file containing the images.
        :param labels_path: Path to the file containing the labels.
        :param number_of_samples: Number of samples from dataset to use.
        """
        self.number_of_samples = number_of_samples
        # load the images and labels from disk
        self.images, self.labels = loadlocal_mnist(images_path=images_path, labels_path=labels_path)
        self.images, self.labels = self.images[:self.number_of_samples], self.labels[:self.number_of_samples]

    def __len__(self):
        # return the number of total samples contained in the dataset
        return self.number_of_samples

    def __getitem__(self, idx):
        # grab the image and label from the current index
        image, label = self.images[idx] / 255., self.labels[idx]

        # return a tuple of the image and its label
        return (image, label)
