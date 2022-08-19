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

    def last_modified(self):


        
class FilesystemDirectory(FilesystemEntry):
    def __init__(self, *args, **kwargs):
        super(FilesystemEntry,self).__init__(*args,**kwargs)

    def canonical_path(self) -> pathlib.Path:
        return self.__canonical_path
