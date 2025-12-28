import datetime

from google.api import annotations_pb2 as _annotations_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class GetCreditBalanceRequest(_message.Message):
    __slots__ = ("user_id",)
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    def __init__(self, user_id: _Optional[int] = ...) -> None: ...

class GetCreditBalanceResponse(_message.Message):
    __slots__ = ("balance", "total_earned", "total_spent", "daily_usage_limit", "today_used")
    BALANCE_FIELD_NUMBER: _ClassVar[int]
    TOTAL_EARNED_FIELD_NUMBER: _ClassVar[int]
    TOTAL_SPENT_FIELD_NUMBER: _ClassVar[int]
    DAILY_USAGE_LIMIT_FIELD_NUMBER: _ClassVar[int]
    TODAY_USED_FIELD_NUMBER: _ClassVar[int]
    balance: int
    total_earned: int
    total_spent: int
    daily_usage_limit: int
    today_used: int
    def __init__(self, balance: _Optional[int] = ..., total_earned: _Optional[int] = ..., total_spent: _Optional[int] = ..., daily_usage_limit: _Optional[int] = ..., today_used: _Optional[int] = ...) -> None: ...

class DeductCreditsRequest(_message.Message):
    __slots__ = ("user_id", "amount", "reason", "idempotency_key", "conversation_id", "message_id")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    AMOUNT_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    IDEMPOTENCY_KEY_FIELD_NUMBER: _ClassVar[int]
    CONVERSATION_ID_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    amount: int
    reason: str
    idempotency_key: str
    conversation_id: int
    message_id: int
    def __init__(self, user_id: _Optional[int] = ..., amount: _Optional[int] = ..., reason: _Optional[str] = ..., idempotency_key: _Optional[str] = ..., conversation_id: _Optional[int] = ..., message_id: _Optional[int] = ...) -> None: ...

class DeductCreditsResponse(_message.Message):
    __slots__ = ("success", "remaining_balance", "transaction_id")
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    REMAINING_BALANCE_FIELD_NUMBER: _ClassVar[int]
    TRANSACTION_ID_FIELD_NUMBER: _ClassVar[int]
    success: bool
    remaining_balance: int
    transaction_id: str
    def __init__(self, success: bool = ..., remaining_balance: _Optional[int] = ..., transaction_id: _Optional[str] = ...) -> None: ...

class AddCreditsRequest(_message.Message):
    __slots__ = ("user_id", "amount", "source", "idempotency_key")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    AMOUNT_FIELD_NUMBER: _ClassVar[int]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    IDEMPOTENCY_KEY_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    amount: int
    source: str
    idempotency_key: str
    def __init__(self, user_id: _Optional[int] = ..., amount: _Optional[int] = ..., source: _Optional[str] = ..., idempotency_key: _Optional[str] = ...) -> None: ...

class AddCreditsResponse(_message.Message):
    __slots__ = ("success", "new_balance", "transaction_id")
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    NEW_BALANCE_FIELD_NUMBER: _ClassVar[int]
    TRANSACTION_ID_FIELD_NUMBER: _ClassVar[int]
    success: bool
    new_balance: int
    transaction_id: str
    def __init__(self, success: bool = ..., new_balance: _Optional[int] = ..., transaction_id: _Optional[str] = ...) -> None: ...

class GetTransactionHistoryRequest(_message.Message):
    __slots__ = ("user_id", "page", "page_size", "type_filter")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    PAGE_FIELD_NUMBER: _ClassVar[int]
    PAGE_SIZE_FIELD_NUMBER: _ClassVar[int]
    TYPE_FILTER_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    page: int
    page_size: int
    type_filter: str
    def __init__(self, user_id: _Optional[int] = ..., page: _Optional[int] = ..., page_size: _Optional[int] = ..., type_filter: _Optional[str] = ...) -> None: ...

class GetTransactionHistoryResponse(_message.Message):
    __slots__ = ("transactions", "total", "page", "page_size")
    TRANSACTIONS_FIELD_NUMBER: _ClassVar[int]
    TOTAL_FIELD_NUMBER: _ClassVar[int]
    PAGE_FIELD_NUMBER: _ClassVar[int]
    PAGE_SIZE_FIELD_NUMBER: _ClassVar[int]
    transactions: _containers.RepeatedCompositeFieldContainer[CreditTransaction]
    total: int
    page: int
    page_size: int
    def __init__(self, transactions: _Optional[_Iterable[_Union[CreditTransaction, _Mapping]]] = ..., total: _Optional[int] = ..., page: _Optional[int] = ..., page_size: _Optional[int] = ...) -> None: ...

class GetUsageStatsRequest(_message.Message):
    __slots__ = ("user_id", "period")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    PERIOD_FIELD_NUMBER: _ClassVar[int]
    user_id: int
    period: str
    def __init__(self, user_id: _Optional[int] = ..., period: _Optional[str] = ...) -> None: ...

class GetUsageStatsResponse(_message.Message):
    __slots__ = ("total_tokens", "total_messages", "total_conversations", "total_credits_used", "daily_usage")
    TOTAL_TOKENS_FIELD_NUMBER: _ClassVar[int]
    TOTAL_MESSAGES_FIELD_NUMBER: _ClassVar[int]
    TOTAL_CONVERSATIONS_FIELD_NUMBER: _ClassVar[int]
    TOTAL_CREDITS_USED_FIELD_NUMBER: _ClassVar[int]
    DAILY_USAGE_FIELD_NUMBER: _ClassVar[int]
    total_tokens: int
    total_messages: int
    total_conversations: int
    total_credits_used: int
    daily_usage: _containers.RepeatedCompositeFieldContainer[DailyUsage]
    def __init__(self, total_tokens: _Optional[int] = ..., total_messages: _Optional[int] = ..., total_conversations: _Optional[int] = ..., total_credits_used: _Optional[int] = ..., daily_usage: _Optional[_Iterable[_Union[DailyUsage, _Mapping]]] = ...) -> None: ...

class CreditTransaction(_message.Message):
    __slots__ = ("id", "user_id", "amount", "type", "reason", "balance_after", "created_at")
    ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    AMOUNT_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    BALANCE_AFTER_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    id: str
    user_id: int
    amount: int
    type: str
    reason: str
    balance_after: int
    created_at: _timestamp_pb2.Timestamp
    def __init__(self, id: _Optional[str] = ..., user_id: _Optional[int] = ..., amount: _Optional[int] = ..., type: _Optional[str] = ..., reason: _Optional[str] = ..., balance_after: _Optional[int] = ..., created_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class DailyUsage(_message.Message):
    __slots__ = ("date", "tokens", "messages", "credits_used")
    DATE_FIELD_NUMBER: _ClassVar[int]
    TOKENS_FIELD_NUMBER: _ClassVar[int]
    MESSAGES_FIELD_NUMBER: _ClassVar[int]
    CREDITS_USED_FIELD_NUMBER: _ClassVar[int]
    date: str
    tokens: int
    messages: int
    credits_used: int
    def __init__(self, date: _Optional[str] = ..., tokens: _Optional[int] = ..., messages: _Optional[int] = ..., credits_used: _Optional[int] = ...) -> None: ...
