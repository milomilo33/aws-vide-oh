use rocket::routes;
use rocket::Build;
use rocket::Rocket;
use crate::connection;
use crate::handler;
use crate::cors::CORS;

use diesel_async::{pooled_connection::bb8::Pool, AsyncPgConnection};
use diesel_async::pooled_connection::AsyncDieselConnectionManager;

pub async fn create_routes(connection_string: &str) -> Result<Rocket<Build>, rocket::Error> {
    println!("before pool");
    let pool = connection::init_pool(&connection_string).await;
    println!("after pool");
    let rocket = Rocket::build()
        .attach(CORS)
        .manage(pool)
        .mount("/dev/api",
            routes![
                handler::show_all_comments_for_video,
                handler::create_comment,
                handler::show_all_reported_comments,
                handler::delete_comment,
                handler::report_comment,
                handler::create_or_update_rating,
                handler::get_rating_for_video,
                handler::get_rating_for_user
            ]
        );
    println!("after build");
    Ok(rocket)
}
