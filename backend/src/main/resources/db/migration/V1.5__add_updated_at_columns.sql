-- Flyway V1.5__add_updated_at_columns.sql
-- 补全所有业务表 updated_at 字段，消除 SQL 引擎缺失字段异常

ALTER TABLE qg_position ADD COLUMN updated_at DATETIME;
ALTER TABLE sq_building ADD COLUMN updated_at DATETIME;
ALTER TABLE sq_inspection ADD COLUMN updated_at DATETIME;
ALTER TABLE sq_room ADD COLUMN updated_at DATETIME;
ALTER TABLE qg_position_apply ADD COLUMN updated_at DATETIME;
ALTER TABLE qg_attendance ADD COLUMN updated_at DATETIME;
ALTER TABLE ty_cultivation_record ADD COLUMN updated_at DATETIME;
ALTER TABLE ty_thought_report ADD COLUMN updated_at DATETIME;
ALTER TABLE ty_political_review ADD COLUMN updated_at DATETIME;
ALTER TABLE ty_development_meeting ADD COLUMN updated_at DATETIME;
ALTER TABLE ty_probationary ADD COLUMN updated_at DATETIME;
ALTER TABLE ty_member_roster ADD COLUMN updated_at DATETIME;
ALTER TABLE st_recruit_apply ADD COLUMN updated_at DATETIME;
