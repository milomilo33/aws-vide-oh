// use std::env;
// use std::ops::Deref;

// use diesel::pg::PgConnection;
// use r2d2;
// use r2d2_diesel::ConnectionManager;
// use rocket::{Outcome, Request, State};
// use rocket::http::Status;
// use rocket::request::{self, FromRequest};

// type Pool = r2d2::Pool<ConnectionManager<PgConnection>>;

// pub fn init_pool() -> Pool {
//     let manager = ConnectionManager::<PgConnection>::new(database_url());
//     Pool::new(manager).expect("db pool")
// }

// fn database_url() -> String {
//     env::var("DATABASE_URL").expect("DATABASE_URL must be set")
// }

// pub struct DbConn(pub r2d2::PooledConnection<ConnectionManager<PgConnection>>);

// impl<'a, 'r> FromRequest<'a, 'r> for DbConn {
//     type Error = ();

//     fn from_request(request: &'a Request<'r>) -> request::Outcome<DbConn, Self::Error> {
//         let pool = request.guard::<State<Pool>>()?;
//         match pool.get() {
//             Ok(conn) => Outcome::Success(DbConn(conn)),
//             Err(_) => Outcome::Failure((Status::ServiceUnavailable, ())),
//         }
//     }
// }

// impl Deref for DbConn {
//     type Target = PgConnection;

//     fn deref(&self) -> &Self::Target {
//         &self.0
//     }
// }

use std::env;
use std::ops::Deref;
use diesel::pg::PgConnection;
use r2d2;
use r2d2_diesel::ConnectionManager;
use rocket::{Outcome, Request, State};
use rocket::http::Status;
use rocket::request::{self, FromRequest};
use rusoto_core::Region;
use rusoto_secretsmanager::{GetSecretValueRequest, SecretsManager, SecretsManagerClient};
use serde::Deserialize;

type Pool = r2d2::Pool<ConnectionManager<PgConnection>>;

#[derive(Deserialize)]
struct DbSecret {
    username: String,
    password: String,
    host: String,
    port: i32,
    dbname: String,
}

pub fn init_pool() -> Pool {
    let db_secret = get_db_secret().expect("Failed to fetch DB secret");
    let connection_string = format!(
        "postgres://{}:{}@{}:{}/{}",
        db_secret.username,
        db_secret.password,
        db_secret.host,
        db_secret.port,
        db_secret.dbname
    );
    let manager = ConnectionManager::<PgConnection>::new(connection_string);
    Pool::new(manager).expect("Failed to create DB pool")
}

fn get_db_secret() -> Result<DbSecret, Box<dyn std::error::Error>> {
    let secret_name = env::var("DB_SECRET_NAME").expect("DB_SECRET_NAME must be set");
    let region = env::var("REGION").expect("REGION must be set");

    let client = SecretsManagerClient::new(Region::default());

    let request = GetSecretValueRequest {
        secret_id: secret_name,
        ..Default::default()
    };

    let result = client.get_secret_value(request).sync()?; // Still using `.sync()` here
    let secret_string = result.secret_string.expect("SecretString is empty");

    let db_secret: DbSecret = serde_json::from_str(&secret_string)?;
    Ok(db_secret)
}

pub struct DbConn(pub r2d2::PooledConnection<ConnectionManager<PgConnection>>);

impl<'a, 'r> FromRequest<'a, 'r> for DbConn {
    type Error = ();

    fn from_request(request: &'a Request<'r>) -> request::Outcome<DbConn, Self::Error> {
        let pool = request.guard::<State<Pool>>()?;
        match pool.get() {
            Ok(conn) => Outcome::Success(DbConn(conn)),
            Err(_) => Outcome::Failure((Status::ServiceUnavailable, ())),
        }
    }
}

impl Deref for DbConn {
    type Target = PgConnection;

    fn deref(&self) -> &Self::Target {
        &self.0
    }
}
