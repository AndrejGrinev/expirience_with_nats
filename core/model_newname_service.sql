/*

    Copyright (c) 2010, 2021 Tender.Pro http://tender.pro.
    [SQL_LICENSE]

    Создание объектов для микросервисов
*/

/* ------------------------------------------------------------------------- */
CREATE SCHEMA newname_service;

CREATE TABLE newname_service.mq_publish (
  id           SERIAL                      NOT NULL
, mq_id        INT                         NOT NULL
, protocol     TEXT                        NOT NULL
, space_name   TEXT                        NOT NULL
, service_name TEXT                        NOT NULL
, entity       TEXT
, method_name  TEXT                        NOT NULL
, timeout      INT                         NOT NULL
, headers      JSONB                       NOT NULL
, body         JSONB                       NOT NULL
, created_at   TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
, direct       BOOLEAN                     NOT NULL DEFAULT FALSE
, reply_to     TEXT                        DEFAULT NULL
, CONSTRAINT fk_newname_mq_publish PRIMARY KEY (mq_id, service_name)
);

COMMENT ON TABLE newname_service.mq_publish               IS 'Задания на публикацию в MQ';
COMMENT ON COLUMN newname_service.mq_publish.id           IS 'ID задания';
COMMENT ON COLUMN newname_service.mq_publish.mq_id        IS 'ID задания в рамках сервиса';
COMMENT ON COLUMN newname_service.mq_publish.protocol     IS 'Протокол';
COMMENT ON COLUMN newname_service.mq_publish.space_name   IS 'Зона';
COMMENT ON COLUMN newname_service.mq_publish.service_name IS 'Сервис-получатель';
COMMENT ON COLUMN newname_service.mq_publish.entity       IS 'Сущность';
COMMENT ON COLUMN newname_service.mq_publish.method_name  IS 'Название метода/события';
COMMENT ON COLUMN newname_service.mq_publish.timeout      IS 'Таймаут. Если 0, то событие';
COMMENT ON COLUMN newname_service.mq_publish.headers      IS 'Заголовок сообщения';
COMMENT ON COLUMN newname_service.mq_publish.body         IS 'Тело сообщения';
COMMENT ON COLUMN newname_service.mq_publish.created_at   IS 'Момент создания';
COMMENT ON COLUMN newname_service.mq_publish.direct       IS 'Создали напрямую? Если да, то нотифай не нужен';
COMMENT ON COLUMN newname_service.mq_publish.reply_to     IS 'Если необходимо ответить, то метка адресата';

/* ------------------------------------------------------------------------- */
CREATE OR REPLACE FUNCTION newname_service.tr_mq_publish_calc_mq_id() RETURNS TRIGGER VOLATILE LANGUAGE 'plpgsql' AS
$_$
  DECLARE
    v_mq_id INTEGER;
  BEGIN
    SELECT INTO v_mq_id COALESCE(MAX(mq_id), 0) FROM newname_service.mq_publish WHERE service_name = NEW.service_name;
    NEW.mq_id := v_mq_id + 1;
    RETURN NEW;
  END;
$_$;

CREATE OR REPLACE FUNCTION newname_service.tr_mq_publish_notify () RETURNS TRIGGER LANGUAGE 'plpgsql' AS
$_$
BEGIN
  PERFORM pg_notify('task_for_newname_service', NEW.id::TEXT);
  RETURN NULL;
END;
$_$;

/* ------------------------------------------------------------------------- */
DROP TRIGGER IF EXISTS set_newname_mq_publish_mq_id on newname_service.mq_publish;
DROP TRIGGER IF EXISTS set_newname_mq_publish_notify on newname_service.mq_publish;

CREATE TRIGGER set_newname_mq_publish_mq_id
  BEFORE INSERT ON newname_service.mq_publish
  FOR EACH ROW
  EXECUTE PROCEDURE newname_service.tr_mq_publish_calc_mq_id()
;

CREATE TRIGGER set_newname_mq_publish_notify
  AFTER INSERT OR UPDATE ON newname_service.mq_publish
  FOR EACH ROW
  WHEN (NOT NEW.direct)
  EXECUTE PROCEDURE newname_service.tr_mq_publish_notify()
;

/* ------------------------------------------------------------------------- */
CREATE TABLE newname_service.mq_receive (
  mq_id        INT                         NOT NULL
, service_name TEXT                        NOT NULL
, received_at  TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
, CONSTRAINT fk_newname_mq_receive PRIMARY KEY (mq_id, service_name)
);

COMMENT ON TABLE newname_service.mq_receive               IS 'Полученные задания от MQ';
COMMENT ON COLUMN newname_service.mq_receive.mq_id        IS 'ID задания в рамках сервиса';
COMMENT ON COLUMN newname_service.mq_receive.service_name IS 'Сервис-отправитель';
COMMENT ON COLUMN newname_service.mq_receive.received_at  IS 'Момент получения';

/* ------------------------------------------------------------------------- */
