#![feature(decl_macro, proc_macro_hygiene)]
#[macro_use]
extern crate diesel;
extern crate dotenv;
#[macro_use]
extern crate rocket;
#[macro_use]
extern crate serde_derive;

use dotenv::dotenv;
use rocket::tokio;

mod schema;
mod connection;
mod models;
mod repository;
mod handler;
mod router;
mod auth;

#[rocket::main]
async fn main() -> Result<(), rocket::Error> {
    dotenv().ok();
    router::create_routes().await?;
    Ok(())
}

// #![feature(proc_macro_hygiene, decl_macro)]

// #[macro_use] extern crate rocket;
// use rocket_lamb::RocketExt;

// #[get("/hello")]
// fn hello() -> &'static str {
//     "Hello, world!"
// }

// fn main() {
//     rocket::ignite()
//         .mount("/dev/api/comments", routes![hello])
//         .lambda() // launch the Rocket as a Lambda
//         .launch();
// }