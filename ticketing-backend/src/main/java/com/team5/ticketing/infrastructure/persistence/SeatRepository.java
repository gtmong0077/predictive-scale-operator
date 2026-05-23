package com.team5.ticketing.infrastructure.persistence;

import java.util.List;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;

import com.team5.ticketing.domain.Seat;

public interface SeatRepository extends JpaRepository<Seat, Long> {
    List<Seat> findByEventIdOrderBySeatNumber(Long eventId);

    Optional<Seat> findByIdAndEventId(Long id, Long eventId);
}