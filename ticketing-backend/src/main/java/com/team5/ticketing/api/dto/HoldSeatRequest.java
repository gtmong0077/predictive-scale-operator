package com.team5.ticketing.api.dto;

import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Positive;

public record HoldSeatRequest(
        @NotNull(message = "userId는 필수입니다.")
        @Positive(message = "userId는 1 이상의 숫자여야 합니다.")
        Long userId
) {
}