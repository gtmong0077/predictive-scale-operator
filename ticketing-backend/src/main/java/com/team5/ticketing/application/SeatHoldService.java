package com.team5.ticketing.application;

import com.team5.ticketing.api.dto.SeatResponse;

public interface SeatHoldService {
    SeatResponse holdSeat(Long eventId, Long seatId, Long userId);
}