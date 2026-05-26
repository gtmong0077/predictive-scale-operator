package com.team5.ticketing.application;

import java.util.List;

import org.springframework.context.annotation.Profile;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.team5.ticketing.api.dto.SeatResponse;
import com.team5.ticketing.domain.Event;
import com.team5.ticketing.domain.Seat;
import com.team5.ticketing.exception.EventNotFoundException;
import com.team5.ticketing.infrastructure.persistence.EventRepository;
import com.team5.ticketing.infrastructure.persistence.SeatRepository;

@Service
@Profile("localdb")
@Transactional
public class DatabaseSeatQueryService implements SeatQueryService {

    private final EventRepository eventRepository;
    private final SeatRepository seatRepository;

    public DatabaseSeatQueryService(EventRepository eventRepository, SeatRepository seatRepository) {
        this.eventRepository = eventRepository;
        this.seatRepository = seatRepository;
    }

    @Override
    public List<SeatResponse> getSeats(Long eventId) {
        Event event = eventRepository.findById(eventId)
                .orElseThrow(() -> new EventNotFoundException(eventId));

        return seatRepository.findByEventIdOrderBySeatNumber(event.getId()).stream()
                .map(this::normalizeAndToResponse)
                .toList();
    }

    private SeatResponse normalizeAndToResponse(Seat seat) {
        if (seat.isExpiredHold()) {
            seat.releaseExpiredHold();
        }

        return new SeatResponse(
                seat.getId(),
                seat.getSeatNumber(),
                seat.getStatus(),
                seat.getHeldBy(),
                seat.getHoldExpiresAt()
        );
    }
}