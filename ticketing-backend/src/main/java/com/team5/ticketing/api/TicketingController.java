package com.team5.ticketing.api;

import java.util.List;
import java.util.Map;

import jakarta.validation.Valid;

import org.springframework.beans.factory.ObjectProvider;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.team5.ticketing.api.dto.HoldSeatRequest;
import com.team5.ticketing.api.dto.SeatResponse;
import com.team5.ticketing.application.SeatHoldService;
import com.team5.ticketing.application.SeatQueryService;
import com.team5.ticketing.application.SeatReserveService;
import com.team5.ticketing.exception.FeatureNotAvailableException;

@RestController
@RequestMapping("/api")
public class TicketingController {

    private final SeatQueryService seatQueryService;
    private final ObjectProvider<SeatHoldService> seatHoldServiceProvider;
    private final ObjectProvider<SeatReserveService> seatReserveServiceProvider;

    public TicketingController(
            SeatQueryService seatQueryService,
            ObjectProvider<SeatHoldService> seatHoldServiceProvider,
            ObjectProvider<SeatReserveService> seatReserveServiceProvider
    ) {
        this.seatQueryService = seatQueryService;
        this.seatHoldServiceProvider = seatHoldServiceProvider;
        this.seatReserveServiceProvider = seatReserveServiceProvider;
    }

    @GetMapping("/hello")
    public Map<String, String> hello() {
        return Map.of(
                "message", "Ticketing backend is running",
                "nextStep", "Seat list comes from MySQL and hold/reserve APIs are available in localdb profile"
        );
    }

    @GetMapping("/events/{eventId}/seats")
    public List<SeatResponse> getSeats(@PathVariable Long eventId) {
        return seatQueryService.getSeats(eventId);
    }

    @PostMapping("/events/{eventId}/seats/{seatId}/hold")
    public SeatResponse holdSeat(
            @PathVariable Long eventId,
            @PathVariable Long seatId,
            @Valid @RequestBody HoldSeatRequest request
    ) {
        SeatHoldService seatHoldService = seatHoldServiceProvider.getIfAvailable();
        if (seatHoldService == null) {
            throw new FeatureNotAvailableException("좌석 선점 API는 localdb 프로필에서만 사용할 수 있습니다.");
        }

        return seatHoldService.holdSeat(eventId, seatId, request.userId());
    }

    @PostMapping("/events/{eventId}/seats/{seatId}/reserve")
    public SeatResponse reserveSeat(
            @PathVariable Long eventId,
            @PathVariable Long seatId,
            @Valid @RequestBody HoldSeatRequest request
    ) {
        SeatReserveService seatReserveService = seatReserveServiceProvider.getIfAvailable();
        if (seatReserveService == null) {
            throw new FeatureNotAvailableException("좌석 예약 확정 API는 localdb 프로필에서만 사용할 수 있습니다.");
        }

        return seatReserveService.reserveSeat(eventId, seatId, request.userId());
    }
}