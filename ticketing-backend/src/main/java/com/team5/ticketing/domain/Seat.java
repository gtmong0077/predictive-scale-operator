package com.team5.ticketing.domain;

import java.time.LocalDateTime;
import java.util.Objects;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.EnumType;
import jakarta.persistence.Enumerated;
import jakarta.persistence.FetchType;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.JoinColumn;
import jakarta.persistence.ManyToOne;
import jakarta.persistence.PrePersist;
import jakarta.persistence.PreUpdate;
import jakarta.persistence.Table;
import jakarta.persistence.UniqueConstraint;
import jakarta.persistence.Version;

@Entity
@Table(
        name = "seats",
        uniqueConstraints = @UniqueConstraint(name = "uq_event_seat", columnNames = {"event_id", "seat_number"})
)
public class Seat {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @ManyToOne(fetch = FetchType.LAZY, optional = false)
    @JoinColumn(name = "event_id", nullable = false)
    private Event event;

    @Column(name = "seat_number", nullable = false, length = 20)
    private String seatNumber;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false, length = 20)
    private SeatStatus status;

    @Column(name = "held_by")
    private Long heldBy;

    @Column(name = "hold_expires_at")
    private LocalDateTime holdExpiresAt;

    @Column(name = "reserved_by")
    private Long reservedBy;

    @Column(name = "reserved_at")
    private LocalDateTime reservedAt;

    @Version
    @Column(nullable = false)
    private Long version;

    @Column(name = "created_at", nullable = false)
    private LocalDateTime createdAt;

    @Column(name = "updated_at", nullable = false)
    private LocalDateTime updatedAt;

    protected Seat() {
    }

    private Seat(Event event, String seatNumber, SeatStatus status) {
        this.event = event;
        this.seatNumber = seatNumber;
        this.status = status;
    }

    public static Seat available(Event event, String seatNumber) {
        return new Seat(event, seatNumber, SeatStatus.AVAILABLE);
    }

    public static Seat held(Event event, String seatNumber, Long heldBy, LocalDateTime holdExpiresAt) {
        Seat seat = new Seat(event, seatNumber, SeatStatus.HELD);
        seat.heldBy = heldBy;
        seat.holdExpiresAt = holdExpiresAt;
        return seat;
    }

    public static Seat reserved(Event event, String seatNumber, Long reservedBy, LocalDateTime reservedAt) {
        Seat seat = new Seat(event, seatNumber, SeatStatus.RESERVED);
        seat.reservedBy = reservedBy;
        seat.reservedAt = reservedAt;
        return seat;
    }

    @PrePersist
    void onCreate() {
        LocalDateTime now = LocalDateTime.now();
        createdAt = now;
        updatedAt = now;
        if (version == null) {
            version = 0L;
        }
    }

    @PreUpdate
    void onUpdate() {
        updatedAt = LocalDateTime.now();
    }

    public boolean isExpiredHold() {
        return isExpiredHoldAt(LocalDateTime.now());
    }

    public boolean isExpiredHoldAt(LocalDateTime time) {
        return status == SeatStatus.HELD
                && holdExpiresAt != null
                && !holdExpiresAt.isAfter(time);
    }

    public boolean canBeHeldAt(LocalDateTime time) {
        return status == SeatStatus.AVAILABLE || isExpiredHoldAt(time);
    }

    public boolean isHeld() {
        return status == SeatStatus.HELD;
    }

    public boolean isHeldBy(Long userId) {
        return isHeld() && Objects.equals(heldBy, userId);
    }

    public boolean isReserved() {
        return status == SeatStatus.RESERVED;
    }

    public void releaseExpiredHold() {
        this.status = SeatStatus.AVAILABLE;
        this.heldBy = null;
        this.holdExpiresAt = null;
    }

    public void hold(Long userId, LocalDateTime expiresAt) {
        this.status = SeatStatus.HELD;
        this.heldBy = userId;
        this.holdExpiresAt = expiresAt;
        this.reservedBy = null;
        this.reservedAt = null;
    }

    public void reserve(Long userId, LocalDateTime reservedTime) {
        this.status = SeatStatus.RESERVED;
        this.heldBy = null;
        this.holdExpiresAt = null;
        this.reservedBy = userId;
        this.reservedAt = reservedTime;
    }

    public Long getId() {
        return id;
    }

    public Event getEvent() {
        return event;
    }

    public String getSeatNumber() {
        return seatNumber;
    }

    public SeatStatus getStatus() {
        return status;
    }

    public Long getHeldBy() {
        return heldBy;
    }

    public LocalDateTime getHoldExpiresAt() {
        return holdExpiresAt;
    }

    public Long getReservedBy() {
        return reservedBy;
    }

    public LocalDateTime getReservedAt() {
        return reservedAt;
    }
}