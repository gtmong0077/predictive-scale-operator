package com.team5.ticketing.application;

import java.util.List;

import com.team5.ticketing.api.dto.SeatResponse;

public interface SeatQueryService {
    List<SeatResponse> getSeats(Long eventId);
}
