package com.team5.ticketing.config;

import java.time.LocalDateTime;
import java.util.List;

import org.springframework.boot.CommandLineRunner;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Profile;

import com.team5.ticketing.domain.Event;
import com.team5.ticketing.domain.Seat;
import com.team5.ticketing.infrastructure.persistence.EventRepository;
import com.team5.ticketing.infrastructure.persistence.SeatRepository;

@Configuration
@Profile("localdb")
public class LocalDbSeedConfig {

    @Bean
    CommandLineRunner seedLocalDb(EventRepository eventRepository, SeatRepository seatRepository) {
        return args -> {
            Event event = eventRepository.findAll().stream()
                    .findFirst()
                    .orElseGet(() -> eventRepository.save(
                            Event.create("Team 5 Ticketing Demo", LocalDateTime.now().plusDays(1))
                    ));

            List<Seat> existingSeats = seatRepository.findByEventIdOrderBySeatNumber(event.getId());
            if (!existingSeats.isEmpty()) {
                return;
            }

            seatRepository.saveAll(List.of(
                    Seat.available(event, "A-1"),
                    Seat.held(event, "A-2", 1001L, LocalDateTime.now().plusMinutes(3)),
                    Seat.reserved(event, "A-3", 2001L, LocalDateTime.now().minusMinutes(5))
            ));
        };
    }
}
