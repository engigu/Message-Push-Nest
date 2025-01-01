import parser from 'cron-parser';


const formatDate = (date) => {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    const seconds = String(date.getSeconds()).padStart(2, '0');
    return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
};

function parseCron(cronExpression) {
    let nextTime = '';
    try {
        const interval = parser.parseExpression(cronExpression);
        const nextDate = interval.next().toDate();
        nextTime = formatDate(nextDate);
    } catch (err) {
        nextTime = '无效的 Cron 表达式';
    }
    return nextTime;
};


export { parseCron };
