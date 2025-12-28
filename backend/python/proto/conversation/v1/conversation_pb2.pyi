import datetime

from google.api import annotations_pb2 as _annotations_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CreateConversationRequest(_message.Message):
    __slots__ = ("user_id", "character_id", "title")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    CHARACTER_ID_FIELD_NUMBER: _ClassVar[int]
    TITLE_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    character_id: int
    title: str
    def __init__(self, user_id: _Optional[int] = ..., character_id: _Optional[int] = ..., title: _Optional[str] = ...) -> None: ...

class CreateConversationResponse(_message.Message):
    __slots__ = ("conversation",)
    CONVERSATION_FIELD_NUMBER: _ClassVar[int]
    conversation: Conversation
    def __init__(self, conversation: _Optional[_Union[Conversation, _Mapping]] = ...) -> None: ...

class GetConversationRequest(_message.Message):
    __slots__ = ("conversation_id",)
    CONVERSATION_ID_FIELD_NUMBER: _ClassVar[int]
    conversation_id: int
    def __init__(self, conversation_id: _Optional[int] = ...) -> None: ...

class GetConversationResponse(_message.Message):
    __slots__ = ("conversation",)
    CONVERSATION_FIELD_NUMBER: _ClassVar[int]
    conversation: Conversation
    def __init__(self, conversation: _Optional[_Union[Conversation, _Mapping]] = ...) -> None: ...

class ListConversationsRequest(_message.Message):
    __slots__ = ("user_id", "page", "page_size", "include_archived")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    PAGE_FIELD_NUMBER: _ClassVar[int]
    PAGE_SIZE_FIELD_NUMBER: _ClassVar[int]
    INCLUDE_ARCHIVED_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    page: int
    page_size: int
    include_archived: bool
    def __init__(self, user_id: _Optional[int] = ..., page: _Optional[int] = ..., page_size: _Optional[int] = ..., include_archived: bool = ...) -> None: ...

class ListConversationsResponse(_message.Message):
    __slots__ = ("conversations", "total", "page", "page_size")
    CONVERSATIONS_FIELD_NUMBER: _ClassVar[int]
    TOTAL_FIELD_NUMBER: _ClassVar[int]
    PAGE_FIELD_NUMBER: _ClassVar[int]
    PAGE_SIZE_FIELD_NUMBER: _ClassVar[int]
    conversations: _containers.RepeatedCompositeFieldContainer[Conversation]
    total: int
    page: int
    page_size: int
    def __init__(self, conversations: _Optional[_Iterable[_Union[Conversation, _Mapping]]] = ..., total: _Optional[int] = ..., page: _Optional[int] = ..., page_size: _Optional[int] = ...) -> None: ...

class SendMessageRequest(_message.Message):
    __slots__ = ("conversation_id", "user_id", "content", "idempotency_key")
    CONVERSATION_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    IDEMPOTENCY_KEY_FIELD_NUMBER: _ClassVar[int]
    conversation_id: int
    user_id: int
    content: str
    idempotency_key: str
    def __init__(self, conversation_id: _Optional[int] = ..., user_id: _Optional[int] = ..., content: _Optional[str] = ..., idempotency_key: _Optional[str] = ...) -> None: ...

class SendMessageResponse(_message.Message):
    __slots__ = ("user_message", "assistant_message", "tokens_used", "credits_deducted", "credits_remaining")
    USER_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    ASSISTANT_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    TOKENS_USED_FIELD_NUMBER: _ClassVar[int]
    CREDITS_DEDUCTED_FIELD_NUMBER: _ClassVar[int]
    CREDITS_REMAINING_FIELD_NUMBER: _ClassVar[int]
    user_message: Message
    assistant_message: Message
    tokens_used: int
    credits_deducted: float
    credits_remaining: float
    def __init__(self, user_message: _Optional[_Union[Message, _Mapping]] = ..., assistant_message: _Optional[_Union[Message, _Mapping]] = ..., tokens_used: _Optional[int] = ..., credits_deducted: _Optional[float] = ..., credits_remaining: _Optional[float] = ...) -> None: ...

class MessageChunk(_message.Message):
    __slots__ = ("conversation_id", "message_id", "delta", "is_final", "tokens_used")
    CONVERSATION_ID_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_ID_FIELD_NUMBER: _ClassVar[int]
    DELTA_FIELD_NUMBER: _ClassVar[int]
    IS_FINAL_FIELD_NUMBER: _ClassVar[int]
    TOKENS_USED_FIELD_NUMBER: _ClassVar[int]
    conversation_id: int
    message_id: int
    delta: str
    is_final: bool
    tokens_used: int
    def __init__(self, conversation_id: _Optional[int] = ..., message_id: _Optional[int] = ..., delta: _Optional[str] = ..., is_final: bool = ..., tokens_used: _Optional[int] = ...) -> None: ...

class GetMessagesRequest(_message.Message):
    __slots__ = ("conversation_id", "limit", "before_message_id")
    CONVERSATION_ID_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    BEFORE_MESSAGE_ID_FIELD_NUMBER: _ClassVar[int]
    conversation_id: int
    limit: int
    before_message_id: int
    def __init__(self, conversation_id: _Optional[int] = ..., limit: _Optional[int] = ..., before_message_id: _Optional[int] = ...) -> None: ...

class GetMessagesResponse(_message.Message):
    __slots__ = ("messages", "has_more", "next_cursor")
    MESSAGES_FIELD_NUMBER: _ClassVar[int]
    HAS_MORE_FIELD_NUMBER: _ClassVar[int]
    NEXT_CURSOR_FIELD_NUMBER: _ClassVar[int]
    messages: _containers.RepeatedCompositeFieldContainer[Message]
    has_more: bool
    next_cursor: str
    def __init__(self, messages: _Optional[_Iterable[_Union[Message, _Mapping]]] = ..., has_more: bool = ..., next_cursor: _Optional[str] = ...) -> None: ...

class ArchiveConversationRequest(_message.Message):
    __slots__ = ("conversation_id", "user_id", "is_archived")
    CONVERSATION_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    IS_ARCHIVED_FIELD_NUMBER: _ClassVar[int]
    conversation_id: int
    user_id: int
    is_archived: bool
    def __init__(self, conversation_id: _Optional[int] = ..., user_id: _Optional[int] = ..., is_archived: bool = ...) -> None: ...

class ArchiveConversationResponse(_message.Message):
    __slots__ = ("success",)
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    success: bool
    def __init__(self, success: bool = ...) -> None: ...

class DeleteConversationRequest(_message.Message):
    __slots__ = ("conversation_id", "user_id")
    CONVERSATION_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    conversation_id: int
    user_id: int
    def __init__(self, conversation_id: _Optional[int] = ..., user_id: _Optional[int] = ...) -> None: ...

class DeleteConversationResponse(_message.Message):
    __slots__ = ("success",)
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    success: bool
    def __init__(self, success: bool = ...) -> None: ...

class Conversation(_message.Message):
    __slots__ = ("id", "user_id", "character_id", "character_name", "character_avatar", "title", "message_count", "token_count", "started_at", "last_message_at", "is_archived")
    ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    CHARACTER_ID_FIELD_NUMBER: _ClassVar[int]
    CHARACTER_NAME_FIELD_NUMBER: _ClassVar[int]
    CHARACTER_AVATAR_FIELD_NUMBER: _ClassVar[int]
    TITLE_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_COUNT_FIELD_NUMBER: _ClassVar[int]
    TOKEN_COUNT_FIELD_NUMBER: _ClassVar[int]
    STARTED_AT_FIELD_NUMBER: _ClassVar[int]
    LAST_MESSAGE_AT_FIELD_NUMBER: _ClassVar[int]
    IS_ARCHIVED_FIELD_NUMBER: _ClassVar[int]
    id: int
    user_id: int
    character_id: int
    character_name: str
    character_avatar: str
    title: str
    message_count: int
    token_count: int
    started_at: _timestamp_pb2.Timestamp
    last_message_at: _timestamp_pb2.Timestamp
    is_archived: bool
    def __init__(self, id: _Optional[int] = ..., user_id: _Optional[int] = ..., character_id: _Optional[int] = ..., character_name: _Optional[str] = ..., character_avatar: _Optional[str] = ..., title: _Optional[str] = ..., message_count: _Optional[int] = ..., token_count: _Optional[int] = ..., started_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., last_message_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., is_archived: bool = ...) -> None: ...

class Message(_message.Message):
    __slots__ = ("id", "conversation_id", "role", "content", "token_count", "metadata", "created_at")
    class MetadataEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    ID_FIELD_NUMBER: _ClassVar[int]
    CONVERSATION_ID_FIELD_NUMBER: _ClassVar[int]
    ROLE_FIELD_NUMBER: _ClassVar[int]
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    TOKEN_COUNT_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    id: int
    conversation_id: int
    role: str
    content: str
    token_count: int
    metadata: _containers.ScalarMap[str, str]
    created_at: _timestamp_pb2.Timestamp
    def __init__(self, id: _Optional[int] = ..., conversation_id: _Optional[int] = ..., role: _Optional[str] = ..., content: _Optional[str] = ..., token_count: _Optional[int] = ..., metadata: _Optional[_Mapping[str, str]] = ..., created_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...
