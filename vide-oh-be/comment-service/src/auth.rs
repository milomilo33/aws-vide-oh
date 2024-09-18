use jwt::{Token, Header, Unverified};
use rocket::outcome::Outcome;
use rocket::http::Status;
use rocket::request::{self, Request, FromRequest};

#[derive(Serialize, Deserialize)]
pub struct MyJWTClaims {
    pub email: String,
    pub role: String,
    pub exp: i64,
}

#[rocket::async_trait]
impl<'r> FromRequest<'r> for MyJWTClaims {
    type Error = Status;

    async fn from_request(request: &'r Request<'_>) -> request::Outcome<Self, Self::Error> {
        let token_strings: Vec<_> = request.headers().get("Authorization").collect();

        match token_strings.len() {
            0 => Outcome::Error((Status::Unauthorized, Status::Unauthorized)),  // No Authorization header
            1 => {
                let parsed_token: Result<Token<Header, MyJWTClaims, Unverified>, jwt::Error> = 
                    Token::parse_unverified(token_strings[0]);

                match parsed_token {
                    Ok(token) => {
                        let my_claims = MyJWTClaims {
                            email: token.claims().email.clone(),
                            role: token.claims().role.clone(),
                            exp: token.claims().exp,
                        };
                        println!("Role: {}", my_claims.role);  // Debug print of the role
                        Outcome::Success(my_claims)  // Successfully parsed claims
                    }
                    Err(_) => Outcome::Error((Status::Unauthorized, Status::Unauthorized)),  // Failed to parse token
                }
            },
            _ => Outcome::Error((Status::BadRequest, Status::Unauthorized)),  // Multiple Authorization headers found
        }
    }
}

