
import pathlib
import datetime 



# a file. it can be compared using its last modified date
class Path:
    # returns a pathlib.Path for a file, using our .path() function if
    # available
    @staticmethod
    def __get_path(pathlike) -> pathlib.Path :
        if type(pathlike) is pathlib.Path:
            return pathlike
        elif issubclass(pathlike, Path):
            return  pathlike.path()
        else:
            return pathlib.Path(pathlike)


    # a pathlib.Path path if available
    def path(self) -> pathlib.Path:
        raise NotImplementedError()
    def last_modified(self) -> datetime.datetime:
        raise NotImplementedError()


    def __lt__(self, other):
        return self.last_modified() < other.last_modified()



