
import pathlib
from ..Path import Path
from ..file import *
from typing import Dict,Iterable,Tuple,List,Iterator
from datetime import datetime

from .Entry import Entry


class File(Entry):
    # presents the canonical name for this entry
    def data(self) -> bytes:
        raise NotImplementedError()

    def update_to_path(self, path: Path) -> bytes:
        with fd as open(path.path(), "wb"):
            fd.write(self.data())




    def update_from_path(self, path: Path) -> bytes:
        with fd as open(path.path()):
            self.update_from_data(fd.read())

        raise NotImplementedError()

    # updates a file based on binary data
    def update_from_data(self, data: bytes) -> bytes:
        raise NotImplementedError()

