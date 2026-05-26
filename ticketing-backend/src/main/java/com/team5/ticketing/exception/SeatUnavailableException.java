package com.team5.ticketing.exception;

import com.team5.ticketing.domain.SeatStatus;

public class SeatUnavailableException extends RuntimeException {

    public SeatUnavailableException(String seatNumber, SeatStatus status) {
        super("좌석 " + seatNumber + "은 현재 " + status + " 상태라서 선점할 수 없습니다.");
    }
}