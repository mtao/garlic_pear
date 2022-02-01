import pathlib
from .path import Path
from .file import *
from typing import List



# a file
class Directory(Path):

    def entries(self) -> List[Path]:
        return list(_ for _ in self.entries_iter())

    def entries_iter(self):
        raise NotImplementedError()

    def all_entries(self) -> List[Path]:
        return list(_ for _ in self.all_entries_iter())

    def all_entries_iter(self):
        raise NotImplementedError()




# defers most tasks to its source's implementation
class SourceDirectory(Directory):
    def __init__(self, source, source_path):
        self.__source = source
        self.__source_path = source_path

    def __str__(self):
        return f"SourceDirectory[{self.__source_path}]"
    # expects a relative path below this path
    def get_child_entry(self, child_path):
        return self.__source.child_entry(self.__source_path / child_path)

    def last_modified(self):
        return self.__source.last_modified(self.__source_path)

# a representation of a directory tracked
class TrackedDirectory(Directory):
    def __init__(self, path):
        self.__path = pathlib.Path(path)
        assert(self.__path.is_dir())

    def __str__(self):
        return f"TrackedDirectory[{self.__path}]"

    def path(self) -> pathlib.Path:
        return self.__path

    def relative_path(self,path) -> str:
        return path.relative_to(self.path())

    def last_modified(self):
        return self.path().stat().st_mtime

    # expects a relative path below this path
    def get_child_entry(self, child_path):
        cp = self.path() / child_path
        if not cp.exists():
            raise FileNotFoundError(cp)
        elif cp.is_dir():
            return TrackedDirectory(cp)
        elif cp.is_file():
            return TrackedFile(cp)
        else:
            raise RuntimeError("unknown file type")



    def entries_iter(self):
        for p in self.path().iterdir():
            yield self.get_child_entry(self.relative_path(p))
    

    def all_entries_iter(self):
        for p in self.path().iterdir():
            e = self.get_child_entry(self.relative_path(p))
            yield e
            if type(e) is TrackedDirectory:
                for e in e.all_entries_iter():
                    yield e

