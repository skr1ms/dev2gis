import logging
import os
from pythonjsonlogger import jsonlogger


class CustomJsonFormatter(jsonlogger.JsonFormatter):
    def add_fields(self, log_record, record, message_dict):
        super(CustomJsonFormatter, self).add_fields(
            log_record, record, message_dict)
        log_record['service'] = 'uav-heightmap'
        log_record['version'] = '1.0.0'
        log_record['environment'] = os.getenv('ENVIRONMENT', 'production')
        log_record['level'] = record.levelname.lower()
        if not log_record.get('timestamp'):
            log_record['time'] = record.created


def setup_logger(name=__name__, level=None):
    if level is None:
        level = os.getenv('LOG_LEVEL', 'INFO').upper()

    logger = logging.getLogger(name)
    logger.setLevel(getattr(logging, level, logging.INFO))

    if not logger.handlers:
        handler = logging.StreamHandler()
        formatter = CustomJsonFormatter(
            '%(time)s %(level)s %(message)s',
            rename_fields={'time': 'time',
                           'level': 'level', 'message': 'message'}
        )
        handler.setFormatter(formatter)
        logger.addHandler(handler)

    return logger


logger = setup_logger('heightmap_service')
