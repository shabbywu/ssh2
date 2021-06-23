# -*- coding: utf-8 -*-
class AttrErrorWhenParsing(Exception):
    pass


class AttrNotDefind(AttrErrorWhenParsing):
    pass


class AttrNotFound(AttrErrorWhenParsing):
    pass


class ImportFromStringError(Exception):
    """Error raise by `import_from_string`"""
