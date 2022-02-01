from ..directory import TrackedDirectory,SourceDirectory
from ..file import SourceFile
from .source import Source
import pathlib




class DirectorySource(Source):
    def __init__(self, path: pathlib.Path):
        path = pathlib.Path(path)
        assert path.is_dir()
        self.__directory = TrackedDirectory(path)

    # convenience function so root access is normalized
    def get_dir_entry(self, path = None):
        if path is None:
            return self.__directory
        else:
            cp = child_path.relative_to(self.__directory.path())
            return self.__directory.get_child_entry(cp)


    # expects a relative path below this path
    def get_child_entry(self, child_path: pathlib.Path):
        cp = self.__directory.relative_path(child_path)

        if not child_path.exists():
            raise FileNotFoundError(child_path)
        elif child_path.is_dir():
            return SourceDirectory(self,cp)
        elif child_path.is_file():
            return SourceFile(self,cp)
        else:
            raise RuntimeError("unknown file type")

    def entries_iter(self, path= None):
        for e in self.get_dir_entry(path).entries_iter():
            yield self.get_child_entry(e.path())

    def all_entries_iter(self, path= None):
        for e in self.get_dir_entry(path).all_entries_iter():
            r = self.get_child_entry(e.path())
            yield r
            if type(r) is TrackedDirectory:
                for e in r.all_entries_iter():
                    yield e
                

    def last_modified(self, entry):
        return self.get_entry(path).last_modified()

