import datetime

from google.api import annotations_pb2 as _annotations_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RegisterRequest(_message.Message):
    __slots__ = ("username", "email", "password", "phone", "verification_code")
    USERNAME_FIELD_NUMBER: _ClassVar[int]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    PASSWORD_FIELD_NUMBER: _ClassVar[int]
    PHONE_FIELD_NUMBER: _ClassVar[int]
    VERIFICATION_CODE_FIELD_NUMBER: _ClassVar[int]
    username: str
    email: str
    password: str
    phone: str
    verification_code: str
    def __init__(self, username: _Optional[str] = ..., email: _Optional[str] = ..., password: _Optional[str] = ..., phone: _Optional[str] = ..., verification_code: _Optional[str] = ...) -> None: ...

class RegisterResponse(_message.Message):
    __slots__ = ("user_id", "access_token", "refresh_token", "expires_in", "initial_credits")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    ACCESS_TOKEN_FIELD_NUMBER: _ClassVar[int]
    REFRESH_TOKEN_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_IN_FIELD_NUMBER: _ClassVar[int]
    INITIAL_CREDITS_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    access_token: str
    refresh_token: str
    expires_in: int
    initial_credits: int
    def __init__(self, user_id: _Optional[int] = ..., access_token: _Optional[str] = ..., refresh_token: _Optional[str] = ..., expires_in: _Optional[int] = ..., initial_credits: _Optional[int] = ...) -> None: ...

class LoginRequest(_message.Message):
    __slots__ = ("email", "password")
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    PASSWORD_FIELD_NUMBER: _ClassVar[int]
    email: str
    password: str
    def __init__(self, email: _Optional[str] = ..., password: _Optional[str] = ...) -> None: ...

class LoginResponse(_message.Message):
    __slots__ = ("user_id", "access_token", "refresh_token", "expires_in", "user")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    ACCESS_TOKEN_FIELD_NUMBER: _ClassVar[int]
    REFRESH_TOKEN_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_IN_FIELD_NUMBER: _ClassVar[int]
    USER_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    access_token: str
    refresh_token: str
    expires_in: int
    user: User
    def __init__(self, user_id: _Optional[int] = ..., access_token: _Optional[str] = ..., refresh_token: _Optional[str] = ..., expires_in: _Optional[int] = ..., user: _Optional[_Union[User, _Mapping]] = ...) -> None: ...

class RefreshTokenRequest(_message.Message):
    __slots__ = ("refresh_token",)
    REFRESH_TOKEN_FIELD_NUMBER: _ClassVar[int]
    refresh_token: str
    def __init__(self, refresh_token: _Optional[str] = ...) -> None: ...

class RefreshTokenResponse(_message.Message):
    __slots__ = ("access_token", "refresh_token", "expires_in")
    ACCESS_TOKEN_FIELD_NUMBER: _ClassVar[int]
    REFRESH_TOKEN_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_IN_FIELD_NUMBER: _ClassVar[int]
    access_token: str
    refresh_token: str
    expires_in: int
    def __init__(self, access_token: _Optional[str] = ..., refresh_token: _Optional[str] = ..., expires_in: _Optional[int] = ...) -> None: ...

class GetUserRequest(_message.Message):
    __slots__ = ("user_id",)
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    def __init__(self, user_id: _Optional[int] = ...) -> None: ...

class GetUserResponse(_message.Message):
    __slots__ = ("user", "profile")
    USER_FIELD_NUMBER: _ClassVar[int]
    PROFILE_FIELD_NUMBER: _ClassVar[int]
    user: User
    profile: UserProfile
    def __init__(self, user: _Optional[_Union[User, _Mapping]] = ..., profile: _Optional[_Union[UserProfile, _Mapping]] = ...) -> None: ...

class UpdateProfileRequest(_message.Message):
    __slots__ = ("user_id", "username", "avatar_url", "profile")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    USERNAME_FIELD_NUMBER: _ClassVar[int]
    AVATAR_URL_FIELD_NUMBER: _ClassVar[int]
    PROFILE_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    username: str
    avatar_url: str
    profile: UserProfile
    def __init__(self, user_id: _Optional[int] = ..., username: _Optional[str] = ..., avatar_url: _Optional[str] = ..., profile: _Optional[_Union[UserProfile, _Mapping]] = ...) -> None: ...

class UpdateProfileResponse(_message.Message):
    __slots__ = ("user", "profile")
    USER_FIELD_NUMBER: _ClassVar[int]
    PROFILE_FIELD_NUMBER: _ClassVar[int]
    user: User
    profile: UserProfile
    def __init__(self, user: _Optional[_Union[User, _Mapping]] = ..., profile: _Optional[_Union[UserProfile, _Mapping]] = ...) -> None: ...

class DeleteAccountRequest(_message.Message):
    __slots__ = ("user_id", "password", "reason")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    PASSWORD_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    password: str
    reason: str
    def __init__(self, user_id: _Optional[int] = ..., password: _Optional[str] = ..., reason: _Optional[str] = ...) -> None: ...

class DeleteAccountResponse(_message.Message):
    __slots__ = ("success", "deletion_scheduled_at", "message")
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    DELETION_SCHEDULED_AT_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    success: bool
    deletion_scheduled_at: _timestamp_pb2.Timestamp
    message: str
    def __init__(self, success: bool = ..., deletion_scheduled_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., message: _Optional[str] = ...) -> None: ...

class RecoverAccountRequest(_message.Message):
    __slots__ = ("user_id", "verification_code")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    VERIFICATION_CODE_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    verification_code: str
    def __init__(self, user_id: _Optional[int] = ..., verification_code: _Optional[str] = ...) -> None: ...

class RecoverAccountResponse(_message.Message):
    __slots__ = ("success", "user", "message")
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    USER_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    success: bool
    user: User
    message: str
    def __init__(self, success: bool = ..., user: _Optional[_Union[User, _Mapping]] = ..., message: _Optional[str] = ...) -> None: ...

class User(_message.Message):
    __slots__ = ("id", "username", "email", "phone", "avatar_url", "created_at", "last_login_at", "is_deleted", "deletion_scheduled_at")
    ID_FIELD_NUMBER: _ClassVar[int]
    USERNAME_FIELD_NUMBER: _ClassVar[int]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    PHONE_FIELD_NUMBER: _ClassVar[int]
    AVATAR_URL_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    LAST_LOGIN_AT_FIELD_NUMBER: _ClassVar[int]
    IS_DELETED_FIELD_NUMBER: _ClassVar[int]
    DELETION_SCHEDULED_AT_FIELD_NUMBER: _ClassVar[int]
    id: int
    username: str
    email: str
    phone: str
    avatar_url: str
    created_at: _timestamp_pb2.Timestamp
    last_login_at: _timestamp_pb2.Timestamp
    is_deleted: bool
    deletion_scheduled_at: _timestamp_pb2.Timestamp
    def __init__(self, id: _Optional[int] = ..., username: _Optional[str] = ..., email: _Optional[str] = ..., phone: _Optional[str] = ..., avatar_url: _Optional[str] = ..., created_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., last_login_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., is_deleted: bool = ..., deletion_scheduled_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class UserProfile(_message.Message):
    __slots__ = ("user_id", "full_name", "gender", "birth_date", "location", "interests", "occupation", "bio")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    FULL_NAME_FIELD_NUMBER: _ClassVar[int]
    GENDER_FIELD_NUMBER: _ClassVar[int]
    BIRTH_DATE_FIELD_NUMBER: _ClassVar[int]
    LOCATION_FIELD_NUMBER: _ClassVar[int]
    INTERESTS_FIELD_NUMBER: _ClassVar[int]
    OCCUPATION_FIELD_NUMBER: _ClassVar[int]
    BIO_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    full_name: str
    gender: str
    birth_date: str
    location: Location
    interests: _containers.RepeatedScalarFieldContainer[str]
    occupation: str
    bio: str
    def __init__(self, user_id: _Optional[int] = ..., full_name: _Optional[str] = ..., gender: _Optional[str] = ..., birth_date: _Optional[str] = ..., location: _Optional[_Union[Location, _Mapping]] = ..., interests: _Optional[_Iterable[str]] = ..., occupation: _Optional[str] = ..., bio: _Optional[str] = ...) -> None: ...

class Location(_message.Message):
    __slots__ = ("country", "city", "timezone")
    COUNTRY_FIELD_NUMBER: _ClassVar[int]
    CITY_FIELD_NUMBER: _ClassVar[int]
    TIMEZONE_FIELD_NUMBER: _ClassVar[int]
    country: str
    city: str
    timezone: str
    def __init__(self, country: _Optional[str] = ..., city: _Optional[str] = ..., timezone: _Optional[str] = ...) -> None: ...
