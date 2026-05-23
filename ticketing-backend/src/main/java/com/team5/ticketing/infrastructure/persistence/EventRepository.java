package com.team5.ticketing.infrastructure.persistence;

import org.springframework.data.jpa.repository.JpaRepository;

import com.team5.ticketing.domain.Event;

public interface EventRepository extends JpaRepository<Event, Long> {
}
