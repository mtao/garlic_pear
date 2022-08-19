import pathlib
from ..path import Path
from ..file import *
from typing import Dict,Iterable,Tuple,List,Iterator


# a collection entry is either a file or directory that we would like to track
class CollectionEntry:

    # presents the canonical name for this entry
    def canonical_name(self):
        raise NotImplementedError()
    # presents the last time an entry was modified
    def last_modified(self):
        raise NotImplementedError()

# a collection of entries (files/directories) that must be tracked

class Collection:
    def entries(self) -> Iterable[CollectionEntry]:
        raise NotImplementedError()

    def entries_list(self) -> List[CollectionEntry]:
        return list(_ for _ in self.entries())

    def all_entries(self) -> Iterator[Tuple[pathlib.Path,CollectionEntry]]:
        return map(lambda x: (x.canonical_name(),x), self.entries())

    def all_entries_dict(self) -> Dict[pathlib.Path, CollectionEntry]:
        return {canon:entry for canon,entry in self.all_entries()}



