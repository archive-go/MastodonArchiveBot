module.exports = {
    apps: [
        {
            script: 'chmod +rwx ./bot && ./bot',
            watch: '.',
            instances: 1,
            ignore_watch: ['logs'],
            log_date_format: 'YYYY-MM-DD HH:mm:ss', //设置日志的时间格式
            log_type: 'json', //输出的日志信息为json格式
            error_file: './logs/error.log', //设置标准错误流日志要写入到哪个文件,代码错误可在此文件查找
            out_file: './logs/console.log', //设置标准输出流日志要写入到哪个文件,如应用的console.log()
            pid_file: './logs/pid.log', //设置pid要写入到哪个文件
        },
    ],
};
