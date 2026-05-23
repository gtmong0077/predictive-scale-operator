package com.team5.ticketing.application;

import java.time.LocalDateTime;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Profile;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.team5.ticketing.api.dto.SeatResponse;
import com.team5.ticketing.domain.Seat;
import com.team5.ticketing.exception.EventNotFoundException;
import com.team5.ticketing.exception.SeatNotFoundException;
import com.team5.ticketing.exception.SeatUnavailableException;
import com.team5.ticketing.infrastructure.persistence.EventRepository;
import com.team5.ticketing.infrastructure.persistence.SeatRepository;

@Service
@Profile("localdb")
@Transactional
public class DatabaseSeatHoldService implements SeatHoldService {

    private final EventRepository eventRepository;
    private final SeatRepository seatRepository;
    private final RedisSeatLockService redisSeatLockService;
    private final long holdDurationMinutes;

    public DatabaseSeatHoldService(
            EventRepository eventRepository,
            SeatRepository seatRepository,
            RedisSeatLockService redisSeatLockService,
            @Value("${ticketing.hold-duration-minutes:3}") long holdDurationMinutes
    ) {
        this.eventRepository = eventRepository;
        this.seatRepository = seatRepository;
        this.redisSeatLockService = redisSeatLockService;
        this.holdDurationMinutes = holdDurationMinutes;
    }

    @Override
    public SeatResponse holdSeat(Long eventId, Long seatId, Long userId) {
        eventRepository.findById(eventId)
                .orElseThrow(() -> new EventNotFoundException(eventId));

        return redisSeatLockService.executeWithSeatLock(eventId, seatId, () -> {
            Seat seat = seatRepository.findByIdAndEventId(seatId, eventId)
                    .orElseThrow(() -> new SeatNotFoundException(eventId, seatId));

            LocalDateTime now = LocalDateTime.now();
            if (!seat.canBeHeldAt(now)) {
                throw new SeatUnavailableException(seat.getSeatNumber(), seat.getStatus());
            }

            seat.hold(userId, now.plusMinutes(holdDurationMinutes));
            Seat savedSeat = seatRepository.save(seat);
            return new SeatResponse(
                    savedSeat.getId(),
                    savedSeat.getSeatNumber(),
                    savedSeat.getStatus(),
                    savedSeat.getHeldBy(),
                    savedSeat.getHoldExpiresAt()
            );
        });
    }
}