package com.team5.ticketing.application;

import java.time.LocalDateTime;

import org.springframework.context.annotation.Profile;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.team5.ticketing.api.dto.SeatResponse;
import com.team5.ticketing.domain.Seat;
import com.team5.ticketing.exception.EventNotFoundException;
import com.team5.ticketing.exception.SeatNotFoundException;
import com.team5.ticketing.exception.SeatReservationConflictException;
import com.team5.ticketing.infrastructure.persistence.EventRepository;
import com.team5.ticketing.infrastructure.persistence.SeatRepository;

@Service
@Profile("localdb")
@Transactional
public class DatabaseSeatReserveService implements SeatReserveService {

    private final EventRepository eventRepository;
    private final SeatRepository seatRepository;
    private final RedisSeatLockService redisSeatLockService;

    public DatabaseSeatReserveService(
            EventRepository eventRepository,
            SeatRepository seatRepository,
            RedisSeatLockService redisSeatLockService
    ) {
        this.eventRepository = eventRepository;
        this.seatRepository = seatRepository;
        this.redisSeatLockService = redisSeatLockService;
    }

    @Override
    public SeatResponse reserveSeat(Long eventId, Long seatId, Long userId) {
        eventRepository.findById(eventId)
                .orElseThrow(() -> new EventNotFoundException(eventId));

        return redisSeatLockService.executeWithSeatLock(eventId, seatId, () -> {
            Seat seat = seatRepository.findByIdAndEventId(seatId, eventId)
                    .orElseThrow(() -> new SeatNotFoundException(eventId, seatId));

            LocalDateTime now = LocalDateTime.now();
            if (seat.isReserved()) {
                throw new SeatReservationConflictException(
                        "좌석 " + seat.getSeatNumber() + "은 이미 예약 완료되었습니다."
                );
            }

            if (!seat.isHeld()) {
                throw new SeatReservationConflictException(
                        "좌석 " + seat.getSeatNumber() + "은 아직 선점되지 않아 예약 확정할 수 없습니다."
                );
            }

            if (seat.isExpiredHoldAt(now)) {
                seat.releaseExpiredHold();
                throw new SeatReservationConflictException(
                        "좌석 " + seat.getSeatNumber() + "의 선점 시간이 만료되어 예약 확정할 수 없습니다."
                );
            }

            if (!seat.isHeldBy(userId)) {
                throw new SeatReservationConflictException(
                        "좌석 " + seat.getSeatNumber() + "은 다른 사용자가 선점 중이라 예약 확정할 수 없습니다."
                );
            }

            seat.reserve(userId, now);
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