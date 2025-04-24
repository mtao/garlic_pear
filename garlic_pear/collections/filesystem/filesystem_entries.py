from ..Collection import Entry as CollectionEntry
import pathlib




class Entry(CollectionEntry):
    # pass in a path and either a canonical path or a root path
    # if the root path is provided then the relative path is the canonical path
    def __init__(self, path: pathlib.Path, root: pathlib.Path = pathlib.Path("."), canonical_path:pathlib.Path = None):

        self.__path = path
        if canonical_path is not None:
            self.__canonical_path = canonical_path
        else:
            self.__canonical_path = path.relative_to(root)

    def path(self) -> pathlib.Path:
        return self.__path
    def canonical_path(self) -> pathlib.Path:
        return self.__canonical_path

    # returns the last time the entity was modified
    # for files this should be mtime, for directories it should be the newest mtime contained within it
    def last_modified(self):
        if not self.path().exists():
            raise FileNotFoundError(cp)
        else:
            return self.path().stat().st_mtime

    def is_directory(self):
        return False


        
