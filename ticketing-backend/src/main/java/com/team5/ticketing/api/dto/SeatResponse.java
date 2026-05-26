package com.team5.ticketing.api.dto;

import java.time.LocalDateTime;

import com.team5.ticketing.domain.SeatStatus;

public record SeatResponse(
        Long seatId,
        String seatNumber,
        SeatStatus status,
        Long heldBy,
        LocalDateTime holdExpiresAt
) {
}
