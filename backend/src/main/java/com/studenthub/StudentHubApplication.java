package com.studenthub;

import org.mybatis.spring.annotation.MapperScan;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication
@MapperScan("com.studenthub.modules.**.mapper")
@EnableScheduling
public class StudentHubApplication {

    public static void main(String[] args) {
        SpringApplication.run(StudentHubApplication.class, args);
    }
}
