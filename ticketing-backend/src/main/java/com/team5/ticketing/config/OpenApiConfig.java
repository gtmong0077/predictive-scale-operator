package com.team5.ticketing.config;

import io.swagger.v3.oas.models.OpenAPI;
import io.swagger.v3.oas.models.info.Info;
import io.swagger.v3.oas.models.info.License;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class OpenApiConfig {

    @Bean
    public OpenAPI ticketingOpenApi() {
        return new OpenAPI()
                .info(new Info()
                        .title("Team 5 Ticketing Backend API")
                        .description("좌석 조회, 선점, 예약 확정을 테스트하기 위한 Spring Boot 백엔드 API")
                        .version("v1")
                        .license(new License().name("Team 5 Internal Use")));
    }
}