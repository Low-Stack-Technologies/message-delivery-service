import { createLogger, format, transports, type Logger } from 'winston'

export default class Log {
  private static instance: Logger

  public static initialize() {
    Log.instance = createLogger({
      level: 'debug',
      format: format.combine(
        format.timestamp({
          format: 'YYYY-MM-DD HH:mm:ss'
        }),
        format.colorize({
          colors: {
            info: 'blue',
            error: 'red',
            warn: 'yellow',
            debug: 'green',
            verbose: 'cyan'
          }
        }),
        format.simple(),
        format.printf((info) => `${info.timestamp} ${info.level}: ${info.message}`)
      ),
      transports: [
        new transports.File({ filename: 'logs/error.log', level: 'error', format: format.uncolorize() }),
        new transports.File({ filename: 'logs/combined.log', format: format.uncolorize() }),
        new transports.Console({ level: 'debug' })
      ]
    })
  }

  public static info = (message: string) => Log.instance.info(message)
  public static error = (message: string) => Log.instance.error(message)
  public static warn = (message: string) => Log.instance.warn(message)
  public static debug = (message: string) => Log.instance.debug(message)
  public static verbose = (message: string) => Log.instance.verbose(message)
}
