import pathlib
from ..Path import Path
from ..file import *
from typing import Dict,Iterable,Tuple,List,Iterator
from datetime import datetime

from .Entry import Entry


class Collection(Entry):
    def entries(self) -> Iterable[Entry]:
        raise NotImplementedError()


    def all_entries(self) -> Iterator[Tuple[pathlib.Path,CollectionEntry]]:
        return  {x.canonical_name():x for x in self.entries()}



    def entries_deep(self) -> Iterable[Entry]:
        for entry in self.entries():
            if isinstance(entry, Collection):
                for e in entries.entries_deep():
                    yield e
            else:
                yield e



    def all_entries_deep(self) -> Iterator[Tuple[pathlib.Path,CollectionEntry]]:
        return  {x.canonical_name():x for x in self.entries_deep()}


    def last_modified(self) -> datetime:
        return max(map(lambda x: x.last_modified(), self.entries()))


