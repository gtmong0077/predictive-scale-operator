package com.team5.ticketing.api;

import java.util.Map;

import org.springframework.context.support.DefaultMessageSourceResolvable;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.MethodArgumentNotValidException;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;

import com.team5.ticketing.exception.EventNotFoundException;
import com.team5.ticketing.exception.FeatureNotAvailableException;
import com.team5.ticketing.exception.SeatLockException;
import com.team5.ticketing.exception.SeatNotFoundException;
import com.team5.ticketing.exception.SeatReservationConflictException;
import com.team5.ticketing.exception.SeatUnavailableException;

@RestControllerAdvice
public class GlobalExceptionHandler {

    @ExceptionHandler(EventNotFoundException.class)
    public ResponseEntity<Map<String, String>> handleEventNotFound(EventNotFoundException exception) {
        return ResponseEntity.status(HttpStatus.NOT_FOUND)
                .body(Map.of(
                        "code", "EVENT_NOT_FOUND",
                        "message", exception.getMessage()
                ));
    }

    @ExceptionHandler(SeatNotFoundException.class)
    public ResponseEntity<Map<String, String>> handleSeatNotFound(SeatNotFoundException exception) {
        return ResponseEntity.status(HttpStatus.NOT_FOUND)
                .body(Map.of(
                        "code", "SEAT_NOT_FOUND",
                        "message", exception.getMessage()
                ));
    }

    @ExceptionHandler(SeatUnavailableException.class)
    public ResponseEntity<Map<String, String>> handleSeatUnavailable(SeatUnavailableException exception) {
        return ResponseEntity.status(HttpStatus.CONFLICT)
                .body(Map.of(
                        "code", "SEAT_NOT_AVAILABLE",
                        "message", exception.getMessage()
                ));
    }

    @ExceptionHandler(SeatReservationConflictException.class)
    public ResponseEntity<Map<String, String>> handleSeatReservationConflict(SeatReservationConflictException exception) {
        return ResponseEntity.status(HttpStatus.CONFLICT)
                .body(Map.of(
                        "code", "SEAT_RESERVE_CONFLICT",
                        "message", exception.getMessage()
                ));
    }

    @ExceptionHandler(SeatLockException.class)
    public ResponseEntity<Map<String, String>> handleSeatLock(SeatLockException exception) {
        return ResponseEntity.status(HttpStatus.CONFLICT)
                .body(Map.of(
                        "code", "SEAT_LOCKED",
                        "message", exception.getMessage()
                ));
    }

    @ExceptionHandler(FeatureNotAvailableException.class)
    public ResponseEntity<Map<String, String>> handleFeatureNotAvailable(FeatureNotAvailableException exception) {
        return ResponseEntity.status(HttpStatus.NOT_IMPLEMENTED)
                .body(Map.of(
                        "code", "FEATURE_NOT_AVAILABLE",
                        "message", exception.getMessage()
                ));
    }

    @ExceptionHandler(MethodArgumentNotValidException.class)
    public ResponseEntity<Map<String, String>> handleValidation(MethodArgumentNotValidException exception) {
        String message = exception.getBindingResult().getFieldErrors().stream()
                .findFirst()
                .map(DefaultMessageSourceResolvable::getDefaultMessage)
                .orElse("요청 값이 올바르지 않습니다.");

        return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                .body(Map.of(
                        "code", "INVALID_REQUEST",
                        "message", message
                ));
    }
}