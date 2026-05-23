package com.team5.ticketing.exception;

public class SeatReservationConflictException extends RuntimeException {

    public SeatReservationConflictException(String message) {
        super(message);
    }
}