import datetime

from google.api import annotations_pb2 as _annotations_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CreateCharacterRequest(_message.Message):
    __slots__ = ("name", "avatar_url", "description", "system_prompt", "greeting", "tags", "is_public", "creator_id")
    NAME_FIELD_NUMBER: _ClassVar[int]
    AVATAR_URL_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    SYSTEM_PROMPT_FIELD_NUMBER: _ClassVar[int]
    GREETING_FIELD_NUMBER: _ClassVar[int]
    TAGS_FIELD_NUMBER: _ClassVar[int]
    IS_PUBLIC_FIELD_NUMBER: _ClassVar[int]
    CREATOR_ID_FIELD_NUMBER: _ClassVar[int]
    name: str
    avatar_url: str
    description: str
    system_prompt: str
    greeting: str
    tags: _containers.RepeatedScalarFieldContainer[str]
    is_public: bool
    creator_id: int
    def __init__(self, name: _Optional[str] = ..., avatar_url: _Optional[str] = ..., description: _Optional[str] = ..., system_prompt: _Optional[str] = ..., greeting: _Optional[str] = ..., tags: _Optional[_Iterable[str]] = ..., is_public: bool = ..., creator_id: _Optional[int] = ...) -> None: ...

class CreateCharacterResponse(_message.Message):
    __slots__ = ("character",)
    CHARACTER_FIELD_NUMBER: _ClassVar[int]
    character: Character
    def __init__(self, character: _Optional[_Union[Character, _Mapping]] = ...) -> None: ...

class UpdateCharacterRequest(_message.Message):
    __slots__ = ("character_id", "name", "avatar_url", "description", "system_prompt", "greeting", "tags", "is_public")
    CHARACTER_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    AVATAR_URL_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    SYSTEM_PROMPT_FIELD_NUMBER: _ClassVar[int]
    GREETING_FIELD_NUMBER: _ClassVar[int]
    TAGS_FIELD_NUMBER: _ClassVar[int]
    IS_PUBLIC_FIELD_NUMBER: _ClassVar[int]
    character_id: int
    name: str
    avatar_url: str
    description: str
    system_prompt: str
    greeting: str
    tags: _containers.RepeatedScalarFieldContainer[str]
    is_public: bool
    def __init__(self, character_id: _Optional[int] = ..., name: _Optional[str] = ..., avatar_url: _Optional[str] = ..., description: _Optional[str] = ..., system_prompt: _Optional[str] = ..., greeting: _Optional[str] = ..., tags: _Optional[_Iterable[str]] = ..., is_public: bool = ...) -> None: ...

class UpdateCharacterResponse(_message.Message):
    __slots__ = ("character",)
    CHARACTER_FIELD_NUMBER: _ClassVar[int]
    character: Character
    def __init__(self, character: _Optional[_Union[Character, _Mapping]] = ...) -> None: ...

class DeleteCharacterRequest(_message.Message):
    __slots__ = ("character_id",)
    CHARACTER_ID_FIELD_NUMBER: _ClassVar[int]
    character_id: int
    def __init__(self, character_id: _Optional[int] = ...) -> None: ...

class DeleteCharacterResponse(_message.Message):
    __slots__ = ("success",)
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    success: bool
    def __init__(self, success: bool = ...) -> None: ...

class GetCharacterRequest(_message.Message):
    __slots__ = ("character_id",)
    CHARACTER_ID_FIELD_NUMBER: _ClassVar[int]
    character_id: int
    def __init__(self, character_id: _Optional[int] = ...) -> None: ...

class GetCharacterResponse(_message.Message):
    __slots__ = ("character",)
    CHARACTER_FIELD_NUMBER: _ClassVar[int]
    character: Character
    def __init__(self, character: _Optional[_Union[Character, _Mapping]] = ...) -> None: ...

class ListCharactersRequest(_message.Message):
    __slots__ = ("page", "page_size", "category", "search", "only_public", "creator_id")
    PAGE_FIELD_NUMBER: _ClassVar[int]
    PAGE_SIZE_FIELD_NUMBER: _ClassVar[int]
    CATEGORY_FIELD_NUMBER: _ClassVar[int]
    SEARCH_FIELD_NUMBER: _ClassVar[int]
    ONLY_PUBLIC_FIELD_NUMBER: _ClassVar[int]
    CREATOR_ID_FIELD_NUMBER: _ClassVar[int]
    page: int
    page_size: int
    category: str
    search: str
    only_public: bool
    creator_id: int
    def __init__(self, page: _Optional[int] = ..., page_size: _Optional[int] = ..., category: _Optional[str] = ..., search: _Optional[str] = ..., only_public: bool = ..., creator_id: _Optional[int] = ...) -> None: ...

class ListCharactersResponse(_message.Message):
    __slots__ = ("characters", "total", "page", "page_size")
    CHARACTERS_FIELD_NUMBER: _ClassVar[int]
    TOTAL_FIELD_NUMBER: _ClassVar[int]
    PAGE_FIELD_NUMBER: _ClassVar[int]
    PAGE_SIZE_FIELD_NUMBER: _ClassVar[int]
    characters: _containers.RepeatedCompositeFieldContainer[Character]
    total: int
    page: int
    page_size: int
    def __init__(self, characters: _Optional[_Iterable[_Union[Character, _Mapping]]] = ..., total: _Optional[int] = ..., page: _Optional[int] = ..., page_size: _Optional[int] = ...) -> None: ...

class Character(_message.Message):
    __slots__ = ("id", "name", "avatar_url", "description", "system_prompt", "greeting", "tags", "is_public", "is_official", "creator_id", "usage_count", "created_at", "updated_at")
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    AVATAR_URL_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    SYSTEM_PROMPT_FIELD_NUMBER: _ClassVar[int]
    GREETING_FIELD_NUMBER: _ClassVar[int]
    TAGS_FIELD_NUMBER: _ClassVar[int]
    IS_PUBLIC_FIELD_NUMBER: _ClassVar[int]
    IS_OFFICIAL_FIELD_NUMBER: _ClassVar[int]
    CREATOR_ID_FIELD_NUMBER: _ClassVar[int]
    USAGE_COUNT_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    id: int
    name: str
    avatar_url: str
    description: str
    system_prompt: str
    greeting: str
    tags: _containers.RepeatedScalarFieldContainer[str]
    is_public: bool
    is_official: bool
    creator_id: int
    usage_count: int
    created_at: _timestamp_pb2.Timestamp
    updated_at: _timestamp_pb2.Timestamp
    def __init__(self, id: _Optional[int] = ..., name: _Optional[str] = ..., avatar_url: _Optional[str] = ..., description: _Optional[str] = ..., system_prompt: _Optional[str] = ..., greeting: _Optional[str] = ..., tags: _Optional[_Iterable[str]] = ..., is_public: bool = ..., is_official: bool = ..., creator_id: _Optional[int] = ..., usage_count: _Optional[int] = ..., created_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., updated_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...
