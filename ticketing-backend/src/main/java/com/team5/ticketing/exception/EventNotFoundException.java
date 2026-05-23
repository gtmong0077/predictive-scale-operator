package com.team5.ticketing.exception;

public class EventNotFoundException extends RuntimeException {

    public EventNotFoundException(Long eventId) {
        super("Event not found. eventId=" + eventId);
    }
}
