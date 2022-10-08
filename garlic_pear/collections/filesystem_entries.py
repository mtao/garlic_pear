from .collection import CollectionEntry
import pathlib




class FilesystemEntry(CollectionEntry):
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


        
class FilesystemDirectory(FilesystemEntry):
    def __init__(self, *args, **kwargs):
        super(self,FilesystemEntry).__init__(*args,**kwargs)

    def canonical_path(self) -> pathlib.Path:
        return self.__canonical_path

    def last_modified(self):
        if self.path().exists():
            return None
        else:
            return self.path().stat().st_mtime

    def is_directory(self):
        return True 

    def is_empty(self) -> bool:
        return len(self.path().iterdir()) == 0


    # expects a relative path below this path
    def get_child_entry(self, child_path):
        cp = self.path() / child_path
        if not cp.exists():
            raise FileNotFoundError(cp)
        elif cp.is_dir():
            return FilesystemDirectory(cp)
        elif cp.is_file():
            return FilesystemEntry(cp)
        else:
            return FilesystemEntry(cp)

    # all file/directories in the directory
    def entries(self):
        for p in self.path().iterdir():
            yield self.get_child_entry(self.relative_path(p))
    

    # all leaf nodes in the directory
    def leaf_child_entries(self):
        for p in self.path().iterdir():
            e = self.get_child_entry(self.relative_path(p))
            if e.is_directory() and e.is_empty():
                for e in e.child_entries():
                    yield e
            else:
                yield e
