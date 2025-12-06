
function generateRandomString(length) {
    const charSet = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
    let result = '';

    for (let i = 0; i < length; i++) {
        const randomIndex = Math.floor(Math.random() * charSet.length);
        result += charSet.charAt(randomIndex);
    }

    return result;
}

function generateUniqueID() {
    const randomString = generateRandomString(10);
    return randomString;
}

function generateBizUniqueID(flag) {
    const randomString = generateRandomString(10);
    return `${flag}${randomString}`;
}

export { generateUniqueID, generateBizUniqueID };
