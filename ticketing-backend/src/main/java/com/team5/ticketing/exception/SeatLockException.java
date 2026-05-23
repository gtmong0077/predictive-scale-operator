package com.team5.ticketing.exception;

public class SeatLockException extends RuntimeException {

    public SeatLockException(Long seatId) {
        super("좌석 " + seatId + "에 대한 다른 요청이 처리 중입니다. 잠시 후 다시 시도해주세요.");
    }
}