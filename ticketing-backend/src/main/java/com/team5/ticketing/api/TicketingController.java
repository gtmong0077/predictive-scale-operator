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

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;

@RestController
@RequestMapping("/api")
@Tag(name = "Ticketing API", description = "좌석 조회, 선점, 예약 확정 API")
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
    @Operation(summary = "서버 실행 확인", description = "백엔드 서버가 정상 실행 중인지 확인합니다.")
    public Map<String, String> hello() {
        return Map.of(
                "message", "Ticketing backend is running",
                "nextStep", "Seat list, hold, reserve, and Swagger UI are available in localdb profile"
        );
    }

    @GetMapping("/events/{eventId}/seats")
    @Operation(summary = "좌석 목록 조회", description = "특정 이벤트의 좌석 상태 목록을 조회합니다.")
    public List<SeatResponse> getSeats(@PathVariable Long eventId) {
        return seatQueryService.getSeats(eventId);
    }

    @PostMapping("/events/{eventId}/seats/{seatId}/hold")
    @Operation(summary = "좌석 선점", description = "비어 있는 좌석을 3분 동안 선점합니다.")
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
    @Operation(summary = "좌석 예약 확정", description = "선점된 좌석을 같은 사용자로 예약 확정합니다.")
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