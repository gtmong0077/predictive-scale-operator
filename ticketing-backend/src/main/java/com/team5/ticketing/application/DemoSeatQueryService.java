package com.team5.ticketing.application;

import java.time.LocalDateTime;
import java.util.List;

import org.springframework.context.annotation.Profile;
import org.springframework.stereotype.Service;

import com.team5.ticketing.api.dto.SeatResponse;
import com.team5.ticketing.domain.SeatStatus;

@Service
@Profile("demo")
public class DemoSeatQueryService implements SeatQueryService {

    @Override
    public List<SeatResponse> getSeats(Long eventId) {
        LocalDateTime holdExpiresAt = LocalDateTime.now().plusMinutes(3);

        return List.of(
                new SeatResponse(1L, "A-1", SeatStatus.AVAILABLE, null, null),
                new SeatResponse(2L, "A-2", SeatStatus.HELD, 1001L, holdExpiresAt),
                new SeatResponse(3L, "A-3", SeatStatus.RESERVED, null, null)
        );
    }
}
