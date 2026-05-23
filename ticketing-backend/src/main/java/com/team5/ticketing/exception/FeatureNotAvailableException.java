package com.team5.ticketing.exception;

public class FeatureNotAvailableException extends RuntimeException {

    public FeatureNotAvailableException(String message) {
        super(message);
    }
}