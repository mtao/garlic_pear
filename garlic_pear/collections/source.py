from typing import List
from ..path import Path
from ..directory import Directory

class Source(Directory):

    def last_modified(self, entry):
        raise NotImplementedError()


    def canonical_entry_dict(self) -> Dict[pathlib.Path, Path]:
        raise NotImplementedError()

