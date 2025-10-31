type LogLevel = 'debug' | 'info' | 'warn' | 'error';

interface LogData {
  [key: string]: any;
}

class Logger {
  private service = 'dev2gis-frontend';
  private version = '1.0.0';
  private environment = typeof process !== 'undefined' && process.env?.NEXT_PUBLIC_ENVIRONMENT 
    ? process.env.NEXT_PUBLIC_ENVIRONMENT 
    : 'production';

  private log(level: LogLevel, message: string, data?: LogData) {
    const logEntry = {
      time: new Date().toISOString(),
      level,
      service: this.service,
      version: this.version,
      environment: this.environment,
      message,
      ...data,
    };

    const logString = JSON.stringify(logEntry);

    switch (level) {
      case 'error':
        console.error(logString);
        break;
      case 'warn':
        console.warn(logString);
        break;
      case 'debug':
        console.debug(logString);
        break;
      default:
        console.log(logString);
    }
  }

  debug(message: string, data?: LogData) {
    this.log('debug', message, data);
  }

  info(message: string, data?: LogData) {
    this.log('info', message, data);
  }

  warn(message: string, data?: LogData) {
    this.log('warn', message, data);
  }

  error(message: string, data?: LogData) {
    this.log('error', message, data);
  }
}

export default new Logger();

