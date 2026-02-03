"""
Exceptions for CRM Platform SDK.
"""

from typing import Any, Dict, Optional


class CRMError(Exception):
    """Base exception for CRM SDK errors."""

    def __init__(
        self,
        message: str,
        status_code: Optional[int] = None,
        error_code: Optional[str] = None,
        details: Optional[Dict[str, Any]] = None,
    ):
        super().__init__(message)
        self.message = message
        self.status_code = status_code
        self.error_code = error_code
        self.details = details or {}

    def __str__(self) -> str:
        parts = [self.message]
        if self.error_code:
            parts.append(f"[{self.error_code}]")
        if self.status_code:
            parts.append(f"(HTTP {self.status_code})")
        return " ".join(parts)

    def __repr__(self) -> str:
        return (
            f"{self.__class__.__name__}("
            f"message={self.message!r}, "
            f"status_code={self.status_code}, "
            f"error_code={self.error_code!r})"
        )


class AuthenticationError(CRMError):
    """Raised when authentication fails."""

    def __init__(
        self,
        message: str = "Authentication failed",
        error_code: Optional[str] = None,
        details: Optional[Dict[str, Any]] = None,
    ):
        super().__init__(
            message=message,
            status_code=401,
            error_code=error_code or "AUTHENTICATION_ERROR",
            details=details,
        )


class AuthorizationError(CRMError):
    """Raised when user lacks permission for an action."""

    def __init__(
        self,
        message: str = "Permission denied",
        error_code: Optional[str] = None,
        details: Optional[Dict[str, Any]] = None,
    ):
        super().__init__(
            message=message,
            status_code=403,
            error_code=error_code or "AUTHORIZATION_ERROR",
            details=details,
        )


class NotFoundError(CRMError):
    """Raised when a resource is not found."""

    def __init__(
        self,
        message: str = "Resource not found",
        resource_type: Optional[str] = None,
        resource_id: Optional[str] = None,
        details: Optional[Dict[str, Any]] = None,
    ):
        details = details or {}
        if resource_type:
            details["resource_type"] = resource_type
        if resource_id:
            details["resource_id"] = resource_id

        super().__init__(
            message=message,
            status_code=404,
            error_code="NOT_FOUND",
            details=details,
        )


class ValidationError(CRMError):
    """Raised when request validation fails."""

    def __init__(
        self,
        message: str = "Validation failed",
        errors: Optional[Dict[str, Any]] = None,
        details: Optional[Dict[str, Any]] = None,
    ):
        details = details or {}
        if errors:
            details["validation_errors"] = errors

        super().__init__(
            message=message,
            status_code=400,
            error_code="VALIDATION_ERROR",
            details=details,
        )


class ConflictError(CRMError):
    """Raised when there's a conflict (e.g., duplicate resource)."""

    def __init__(
        self,
        message: str = "Resource conflict",
        error_code: Optional[str] = None,
        details: Optional[Dict[str, Any]] = None,
    ):
        super().__init__(
            message=message,
            status_code=409,
            error_code=error_code or "CONFLICT",
            details=details,
        )


class RateLimitError(CRMError):
    """Raised when rate limit is exceeded."""

    def __init__(
        self,
        message: str = "Rate limit exceeded",
        retry_after: Optional[int] = None,
        details: Optional[Dict[str, Any]] = None,
    ):
        details = details or {}
        if retry_after:
            details["retry_after"] = retry_after

        super().__init__(
            message=message,
            status_code=429,
            error_code="RATE_LIMIT_EXCEEDED",
            details=details,
        )
        self.retry_after = retry_after


class ServerError(CRMError):
    """Raised when the server returns a 5xx error."""

    def __init__(
        self,
        message: str = "Server error",
        status_code: int = 500,
        error_code: Optional[str] = None,
        details: Optional[Dict[str, Any]] = None,
    ):
        super().__init__(
            message=message,
            status_code=status_code,
            error_code=error_code or "SERVER_ERROR",
            details=details,
        )


class NetworkError(CRMError):
    """Raised when there's a network connectivity issue."""

    def __init__(
        self,
        message: str = "Network error",
        original_error: Optional[Exception] = None,
        details: Optional[Dict[str, Any]] = None,
    ):
        details = details or {}
        if original_error:
            details["original_error"] = str(original_error)

        super().__init__(
            message=message,
            status_code=None,
            error_code="NETWORK_ERROR",
            details=details,
        )
        self.original_error = original_error


class TimeoutError(CRMError):
    """Raised when a request times out."""

    def __init__(
        self,
        message: str = "Request timed out",
        timeout: Optional[float] = None,
        details: Optional[Dict[str, Any]] = None,
    ):
        details = details or {}
        if timeout:
            details["timeout"] = timeout

        super().__init__(
            message=message,
            status_code=None,
            error_code="TIMEOUT",
            details=details,
        )
        self.timeout = timeout


def raise_for_status(status_code: int, response_data: Dict[str, Any]) -> None:
    """Raise appropriate exception based on status code."""
    error_message = response_data.get("error", response_data.get("message", "Unknown error"))
    error_code = response_data.get("code")
    details = response_data.get("details", {})

    if status_code == 400:
        raise ValidationError(
            message=error_message,
            errors=response_data.get("errors"),
            details=details,
        )
    elif status_code == 401:
        raise AuthenticationError(
            message=error_message,
            error_code=error_code,
            details=details,
        )
    elif status_code == 403:
        raise AuthorizationError(
            message=error_message,
            error_code=error_code,
            details=details,
        )
    elif status_code == 404:
        raise NotFoundError(
            message=error_message,
            details=details,
        )
    elif status_code == 409:
        raise ConflictError(
            message=error_message,
            error_code=error_code,
            details=details,
        )
    elif status_code == 429:
        raise RateLimitError(
            message=error_message,
            retry_after=response_data.get("retry_after"),
            details=details,
        )
    elif status_code >= 500:
        raise ServerError(
            message=error_message,
            status_code=status_code,
            error_code=error_code,
            details=details,
        )
    else:
        raise CRMError(
            message=error_message,
            status_code=status_code,
            error_code=error_code,
            details=details,
        )
