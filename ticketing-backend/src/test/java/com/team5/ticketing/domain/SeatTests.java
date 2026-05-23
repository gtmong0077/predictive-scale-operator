package com.team5.ticketing.domain;

import static org.assertj.core.api.Assertions.assertThat;

import java.time.LocalDateTime;

import org.junit.jupiter.api.Test;

class SeatTests {

    @Test
    void expiredHoldShouldBecomeHoldableAgain() {
        Event event = Event.create("Test Event", LocalDateTime.of(2026, 5, 23, 20, 0));
        Seat seat = Seat.held(event, "A-1", 1001L, LocalDateTime.of(2026, 5, 23, 19, 0));

        LocalDateTime now = LocalDateTime.of(2026, 5, 23, 19, 5);

        assertThat(seat.isExpiredHoldAt(now)).isTrue();
        assertThat(seat.canBeHeldAt(now)).isTrue();
    }

    @Test
    void releaseExpiredHoldShouldResetSeatToAvailable() {
        Event event = Event.create("Test Event", LocalDateTime.of(2026, 5, 23, 20, 0));
        Seat seat = Seat.held(event, "A-1", 1001L, LocalDateTime.of(2026, 5, 23, 19, 0));

        seat.releaseExpiredHold();

        assertThat(seat.getStatus()).isEqualTo(SeatStatus.AVAILABLE);
        assertThat(seat.getHeldBy()).isNull();
        assertThat(seat.getHoldExpiresAt()).isNull();
    }

    @Test
    void holdShouldUpdateSeatState() {
        Event event = Event.create("Test Event", LocalDateTime.of(2026, 5, 23, 20, 0));
        Seat seat = Seat.available(event, "A-2");

        LocalDateTime expiresAt = LocalDateTime.of(2026, 5, 23, 20, 3);
        seat.hold(3001L, expiresAt);

        assertThat(seat.getStatus()).isEqualTo(SeatStatus.HELD);
        assertThat(seat.getHeldBy()).isEqualTo(3001L);
        assertThat(seat.getHoldExpiresAt()).isEqualTo(expiresAt);
    }

    @Test
    void reserveShouldUpdateSeatStateAndClearHoldInfo() {
        Event event = Event.create("Test Event", LocalDateTime.of(2026, 5, 23, 20, 0));
        Seat seat = Seat.held(event, "A-3", 3001L, LocalDateTime.of(2026, 5, 23, 20, 3));

        LocalDateTime reservedAt = LocalDateTime.of(2026, 5, 23, 20, 1);
        seat.reserve(3001L, reservedAt);

        assertThat(seat.getStatus()).isEqualTo(SeatStatus.RESERVED);
        assertThat(seat.getHeldBy()).isNull();
        assertThat(seat.getHoldExpiresAt()).isNull();
        assertThat(seat.getReservedBy()).isEqualTo(3001L);
        assertThat(seat.getReservedAt()).isEqualTo(reservedAt);
    }
}