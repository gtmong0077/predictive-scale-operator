package com.team5.ticketing.application;

import com.team5.ticketing.api.dto.SeatResponse;

public interface SeatReserveService {
    SeatResponse reserveSeat(Long eventId, Long seatId, Long userId);
}