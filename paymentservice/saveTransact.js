const Payment = require("./model/payment");
const con = require("./config/mysql");


async function saveTransact(cardNumber, cardType, amount) {



    payment = new Payment({
        cardNumber,
        cardType,
        amount
    });

    payment.save();
    console.log(" Transaction Saved in database ");

    /* var sql = "INSERT INTO customers (name, address) VALUES ('Company Inc', 'Highway 37')";
    con.query(sql, function(err, result) {
        if (err) throw err;
        console.log("1 record inserted");
    });
*/
}

module.exports = saveTransact;