
import pathlib



# a file
class Path:

    
    @staticmethod
    def __get_path(pathlike) -> pathlib.Path :
        if type(pathlike) is pathlib.Path:
            return path_or_file
        elif issubclass(pathlike, Path):
            return  other_file.path()
        else:
            return pathlib.Path(pathlike)




    def path(self) -> pathlib.Path:
        raise NotImplementedError()
    def last_modified(self):
        raise NotImplementedError()


    def __lt__(self, other):
        return str(self) < str(other)



