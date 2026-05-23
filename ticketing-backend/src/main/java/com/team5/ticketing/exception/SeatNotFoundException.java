package com.team5.ticketing.exception;

public class SeatNotFoundException extends RuntimeException {

    public SeatNotFoundException(Long eventId, Long seatId) {
        super("이벤트 " + eventId + "에서 좌석 " + seatId + "를 찾을 수 없습니다.");
    }
}