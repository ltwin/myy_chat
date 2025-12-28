import datetime

from google.api import annotations_pb2 as _annotations_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class VisibilityScope(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    VISIBILITY_SCOPE_UNSPECIFIED: _ClassVar[VisibilityScope]
    OWNER_PRIVATE: _ClassVar[VisibilityScope]
    PUBLIC: _ClassVar[VisibilityScope]
    VISITOR_PRIVATE: _ClassVar[VisibilityScope]

class MemoryType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    MEMORY_TYPE_UNSPECIFIED: _ClassVar[MemoryType]
    USER_PROFILE: _ClassVar[MemoryType]
    EVENT: _ClassVar[MemoryType]
    EMOTION: _ClassVar[MemoryType]
    CORE: _ClassVar[MemoryType]
    TRIVIA: _ClassVar[MemoryType]
VISIBILITY_SCOPE_UNSPECIFIED: VisibilityScope
OWNER_PRIVATE: VisibilityScope
PUBLIC: VisibilityScope
VISITOR_PRIVATE: VisibilityScope
MEMORY_TYPE_UNSPECIFIED: MemoryType
USER_PROFILE: MemoryType
EVENT: MemoryType
EMOTION: MemoryType
CORE: MemoryType
TRIVIA: MemoryType

class SaveMemoryRequest(_message.Message):
    __slots__ = ("user_id", "character_id", "content", "type", "importance", "structured_data", "embedding", "actor_user_id", "visibility_scope")
    class StructuredDataEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    CHARACTER_ID_FIELD_NUMBER: _ClassVar[int]
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    IMPORTANCE_FIELD_NUMBER: _ClassVar[int]
    STRUCTURED_DATA_FIELD_NUMBER: _ClassVar[int]
    EMBEDDING_FIELD_NUMBER: _ClassVar[int]
    ACTOR_USER_ID_FIELD_NUMBER: _ClassVar[int]
    VISIBILITY_SCOPE_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    character_id: int
    content: str
    type: MemoryType
    importance: int
    structured_data: _containers.ScalarMap[str, str]
    embedding: _containers.RepeatedScalarFieldContainer[float]
    actor_user_id: int
    visibility_scope: VisibilityScope
    def __init__(self, user_id: _Optional[int] = ..., character_id: _Optional[int] = ..., content: _Optional[str] = ..., type: _Optional[_Union[MemoryType, str]] = ..., importance: _Optional[int] = ..., structured_data: _Optional[_Mapping[str, str]] = ..., embedding: _Optional[_Iterable[float]] = ..., actor_user_id: _Optional[int] = ..., visibility_scope: _Optional[_Union[VisibilityScope, str]] = ...) -> None: ...

class SaveMemoryResponse(_message.Message):
    __slots__ = ("memory",)
    MEMORY_FIELD_NUMBER: _ClassVar[int]
    memory: Memory
    def __init__(self, memory: _Optional[_Union[Memory, _Mapping]] = ...) -> None: ...

class RetrieveMemoriesRequest(_message.Message):
    __slots__ = ("user_id", "character_id", "query", "query_embedding", "limit", "min_similarity", "types", "actor_user_id", "allowed_scopes")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    CHARACTER_ID_FIELD_NUMBER: _ClassVar[int]
    QUERY_FIELD_NUMBER: _ClassVar[int]
    QUERY_EMBEDDING_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    MIN_SIMILARITY_FIELD_NUMBER: _ClassVar[int]
    TYPES_FIELD_NUMBER: _ClassVar[int]
    ACTOR_USER_ID_FIELD_NUMBER: _ClassVar[int]
    ALLOWED_SCOPES_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    character_id: int
    query: str
    query_embedding: _containers.RepeatedScalarFieldContainer[float]
    limit: int
    min_similarity: float
    types: _containers.RepeatedScalarFieldContainer[MemoryType]
    actor_user_id: int
    allowed_scopes: _containers.RepeatedScalarFieldContainer[VisibilityScope]
    def __init__(self, user_id: _Optional[int] = ..., character_id: _Optional[int] = ..., query: _Optional[str] = ..., query_embedding: _Optional[_Iterable[float]] = ..., limit: _Optional[int] = ..., min_similarity: _Optional[float] = ..., types: _Optional[_Iterable[_Union[MemoryType, str]]] = ..., actor_user_id: _Optional[int] = ..., allowed_scopes: _Optional[_Iterable[_Union[VisibilityScope, str]]] = ...) -> None: ...

class RetrieveMemoriesResponse(_message.Message):
    __slots__ = ("memories",)
    MEMORIES_FIELD_NUMBER: _ClassVar[int]
    memories: _containers.RepeatedCompositeFieldContainer[MemoryWithScore]
    def __init__(self, memories: _Optional[_Iterable[_Union[MemoryWithScore, _Mapping]]] = ...) -> None: ...

class GetUserPortraitRequest(_message.Message):
    __slots__ = ("user_id", "character_id")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    CHARACTER_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    character_id: int
    def __init__(self, user_id: _Optional[int] = ..., character_id: _Optional[int] = ...) -> None: ...

class GetUserPortraitResponse(_message.Message):
    __slots__ = ("portrait",)
    PORTRAIT_FIELD_NUMBER: _ClassVar[int]
    portrait: UserPortrait
    def __init__(self, portrait: _Optional[_Union[UserPortrait, _Mapping]] = ...) -> None: ...

class UpdateRelationshipRequest(_message.Message):
    __slots__ = ("user_id", "character_id", "closeness_delta", "emotional_bond_updates")
    class EmotionalBondUpdatesEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: int
        def __init__(self, key: _Optional[str] = ..., value: _Optional[int] = ...) -> None: ...
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    CHARACTER_ID_FIELD_NUMBER: _ClassVar[int]
    CLOSENESS_DELTA_FIELD_NUMBER: _ClassVar[int]
    EMOTIONAL_BOND_UPDATES_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    character_id: int
    closeness_delta: float
    emotional_bond_updates: _containers.ScalarMap[str, int]
    def __init__(self, user_id: _Optional[int] = ..., character_id: _Optional[int] = ..., closeness_delta: _Optional[float] = ..., emotional_bond_updates: _Optional[_Mapping[str, int]] = ...) -> None: ...

class UpdateRelationshipResponse(_message.Message):
    __slots__ = ("relationship",)
    RELATIONSHIP_FIELD_NUMBER: _ClassVar[int]
    relationship: Relationship
    def __init__(self, relationship: _Optional[_Union[Relationship, _Mapping]] = ...) -> None: ...

class GetRelationshipRequest(_message.Message):
    __slots__ = ("user_id", "character_id")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    CHARACTER_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    character_id: int
    def __init__(self, user_id: _Optional[int] = ..., character_id: _Optional[int] = ...) -> None: ...

class GetRelationshipResponse(_message.Message):
    __slots__ = ("relationship",)
    RELATIONSHIP_FIELD_NUMBER: _ClassVar[int]
    relationship: Relationship
    def __init__(self, relationship: _Optional[_Union[Relationship, _Mapping]] = ...) -> None: ...

class DeleteMemoryRequest(_message.Message):
    __slots__ = ("memory_id",)
    MEMORY_ID_FIELD_NUMBER: _ClassVar[int]
    memory_id: int
    def __init__(self, memory_id: _Optional[int] = ...) -> None: ...

class DeleteMemoryResponse(_message.Message):
    __slots__ = ("success",)
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    success: bool
    def __init__(self, success: bool = ...) -> None: ...

class DetectContradictionsRequest(_message.Message):
    __slots__ = ("user_id", "character_id", "new_memory")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    CHARACTER_ID_FIELD_NUMBER: _ClassVar[int]
    NEW_MEMORY_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    character_id: int
    new_memory: str
    def __init__(self, user_id: _Optional[int] = ..., character_id: _Optional[int] = ..., new_memory: _Optional[str] = ...) -> None: ...

class DetectContradictionsResponse(_message.Message):
    __slots__ = ("has_contradiction", "contradicted_memories", "resolution")
    HAS_CONTRADICTION_FIELD_NUMBER: _ClassVar[int]
    CONTRADICTED_MEMORIES_FIELD_NUMBER: _ClassVar[int]
    RESOLUTION_FIELD_NUMBER: _ClassVar[int]
    has_contradiction: bool
    contradicted_memories: _containers.RepeatedCompositeFieldContainer[Memory]
    resolution: str
    def __init__(self, has_contradiction: bool = ..., contradicted_memories: _Optional[_Iterable[_Union[Memory, _Mapping]]] = ..., resolution: _Optional[str] = ...) -> None: ...

class ReportAuditEventRequest(_message.Message):
    __slots__ = ("character_id", "owner_user_id", "actor_user_id", "conversation_id", "event_type", "severity", "summary", "metadata")
    class MetadataEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    CHARACTER_ID_FIELD_NUMBER: _ClassVar[int]
    OWNER_USER_ID_FIELD_NUMBER: _ClassVar[int]
    ACTOR_USER_ID_FIELD_NUMBER: _ClassVar[int]
    CONVERSATION_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_TYPE_FIELD_NUMBER: _ClassVar[int]
    SEVERITY_FIELD_NUMBER: _ClassVar[int]
    SUMMARY_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    character_id: int
    owner_user_id: int
    actor_user_id: int
    conversation_id: int
    event_type: str
    severity: str
    summary: str
    metadata: _containers.ScalarMap[str, str]
    def __init__(self, character_id: _Optional[int] = ..., owner_user_id: _Optional[int] = ..., actor_user_id: _Optional[int] = ..., conversation_id: _Optional[int] = ..., event_type: _Optional[str] = ..., severity: _Optional[str] = ..., summary: _Optional[str] = ..., metadata: _Optional[_Mapping[str, str]] = ...) -> None: ...

class ReportAuditEventResponse(_message.Message):
    __slots__ = ("event_id", "success")
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    event_id: int
    success: bool
    def __init__(self, event_id: _Optional[int] = ..., success: bool = ...) -> None: ...

class Memory(_message.Message):
    __slots__ = ("id", "user_id", "character_id", "type", "content", "importance", "structured_data", "access_count", "created_at", "last_accessed_at", "visibility_scope", "actor_user_id")
    class StructuredDataEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    CHARACTER_ID_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    IMPORTANCE_FIELD_NUMBER: _ClassVar[int]
    STRUCTURED_DATA_FIELD_NUMBER: _ClassVar[int]
    ACCESS_COUNT_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    LAST_ACCESSED_AT_FIELD_NUMBER: _ClassVar[int]
    VISIBILITY_SCOPE_FIELD_NUMBER: _ClassVar[int]
    ACTOR_USER_ID_FIELD_NUMBER: _ClassVar[int]
    id: int
    user_id: int
    character_id: int
    type: MemoryType
    content: str
    importance: int
    structured_data: _containers.ScalarMap[str, str]
    access_count: int
    created_at: _timestamp_pb2.Timestamp
    last_accessed_at: _timestamp_pb2.Timestamp
    visibility_scope: VisibilityScope
    actor_user_id: int
    def __init__(self, id: _Optional[int] = ..., user_id: _Optional[int] = ..., character_id: _Optional[int] = ..., type: _Optional[_Union[MemoryType, str]] = ..., content: _Optional[str] = ..., importance: _Optional[int] = ..., structured_data: _Optional[_Mapping[str, str]] = ..., access_count: _Optional[int] = ..., created_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., last_accessed_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., visibility_scope: _Optional[_Union[VisibilityScope, str]] = ..., actor_user_id: _Optional[int] = ...) -> None: ...

class MemoryWithScore(_message.Message):
    __slots__ = ("memory", "similarity", "final_score")
    MEMORY_FIELD_NUMBER: _ClassVar[int]
    SIMILARITY_FIELD_NUMBER: _ClassVar[int]
    FINAL_SCORE_FIELD_NUMBER: _ClassVar[int]
    memory: Memory
    similarity: float
    final_score: float
    def __init__(self, memory: _Optional[_Union[Memory, _Mapping]] = ..., similarity: _Optional[float] = ..., final_score: _Optional[float] = ...) -> None: ...

class UserPortrait(_message.Message):
    __slots__ = ("user_id", "character_id", "full_name", "age", "gender", "interests", "occupation", "personality", "important_facts")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    CHARACTER_ID_FIELD_NUMBER: _ClassVar[int]
    FULL_NAME_FIELD_NUMBER: _ClassVar[int]
    AGE_FIELD_NUMBER: _ClassVar[int]
    GENDER_FIELD_NUMBER: _ClassVar[int]
    INTERESTS_FIELD_NUMBER: _ClassVar[int]
    OCCUPATION_FIELD_NUMBER: _ClassVar[int]
    PERSONALITY_FIELD_NUMBER: _ClassVar[int]
    IMPORTANT_FACTS_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    character_id: int
    full_name: str
    age: int
    gender: str
    interests: _containers.RepeatedScalarFieldContainer[str]
    occupation: str
    personality: str
    important_facts: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, user_id: _Optional[int] = ..., character_id: _Optional[int] = ..., full_name: _Optional[str] = ..., age: _Optional[int] = ..., gender: _Optional[str] = ..., interests: _Optional[_Iterable[str]] = ..., occupation: _Optional[str] = ..., personality: _Optional[str] = ..., important_facts: _Optional[_Iterable[str]] = ...) -> None: ...

class Relationship(_message.Message):
    __slots__ = ("user_id", "character_id", "relation_type", "closeness", "emotional_bond", "interaction_count", "total_conversation_time", "last_interaction_at", "is_owner", "permission_level", "loyalty_lock")
    class EmotionalBondEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: int
        def __init__(self, key: _Optional[str] = ..., value: _Optional[int] = ...) -> None: ...
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    CHARACTER_ID_FIELD_NUMBER: _ClassVar[int]
    RELATION_TYPE_FIELD_NUMBER: _ClassVar[int]
    CLOSENESS_FIELD_NUMBER: _ClassVar[int]
    EMOTIONAL_BOND_FIELD_NUMBER: _ClassVar[int]
    INTERACTION_COUNT_FIELD_NUMBER: _ClassVar[int]
    TOTAL_CONVERSATION_TIME_FIELD_NUMBER: _ClassVar[int]
    LAST_INTERACTION_AT_FIELD_NUMBER: _ClassVar[int]
    IS_OWNER_FIELD_NUMBER: _ClassVar[int]
    PERMISSION_LEVEL_FIELD_NUMBER: _ClassVar[int]
    LOYALTY_LOCK_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    character_id: int
    relation_type: str
    closeness: float
    emotional_bond: _containers.ScalarMap[str, int]
    interaction_count: int
    total_conversation_time: int
    last_interaction_at: _timestamp_pb2.Timestamp
    is_owner: bool
    permission_level: str
    loyalty_lock: bool
    def __init__(self, user_id: _Optional[int] = ..., character_id: _Optional[int] = ..., relation_type: _Optional[str] = ..., closeness: _Optional[float] = ..., emotional_bond: _Optional[_Mapping[str, int]] = ..., interaction_count: _Optional[int] = ..., total_conversation_time: _Optional[int] = ..., last_interaction_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., is_owner: bool = ..., permission_level: _Optional[str] = ..., loyalty_lock: bool = ...) -> None: ...
