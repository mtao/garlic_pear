
import pathlib
from .utils.diff import diff
from .path import Path



# a file
class File(Path):

    def __init__(self, path):
        self.__path = pathlib.Path(path)

    def path(self) -> pathlib.Path:
        raise NotImplementedError()
    
    @staticmethod
    def __get_path(path_or_file) -> pathlib.Path :
        if type(path_or_file) is pathlib.Path:
            return path_or_file
        elif issubclass(other_file, File):
            return  other_file.path()
        else:
            return pathlib.Path(path_or_file)



    def diff(self, other_file):
        other_path = File.__get_path(other_file)
        utils.diff(self.path(),other_path)


    # if a file exists then returns the mtime, otherwise None
    def last_modified(self):
        if self.__path.exists():
            return None
        else:
            return self.path().stat().st_mtime



# defers most tasks to its source's implementation
class SourceFile(File):
    def __init__(self, source, source_path):
        self.__source = source
        self.__source_path = source_path
    def __str__(self):
        return f"SourceFile[{self.__source_path}]"

    def last_modified(self):
        return self.__source.last_modified(self.__source_path)

    def path(self) -> pathlib.Path:
        return self.__path


# a representation of a file tracked
class TrackedFile(File):
    def __init__(self, path):
        self.__path = pathlib.Path(path)
        assert(self.__path.is_file())

    def __str__(self):
        return f"TrackedFile[{self.__path}]"

    def path(self) -> pathlib.Path:
        return self.__path



