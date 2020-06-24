BEGIN;

CREATE TABLE products (
    id bigserial not null primary key,
    name varchar null,
    slug varchar null,
    description varchar null
);

COMMIT;

INSERT INTO products (name, slug, description) VALUES
    ('World of Authcraft', 'world-of-authcraft', 'Battle bugs and protect yourself from invaders while you explore a scary world with no security'),
	('Ocean Explorer', 'ocean-explorer', 'Explore the depths of the sea in this one of a kind underwater experience'),
	('Dinosaur Park', 'dinosaur-park', 'Go back 65 million years in the past and ride a T-Rex'),
	('Cars VR', 'cars-vr', 'Get behind the wheel of the fastest cars in the world.'),
	('Robin Hood', 'robin-hood', 'Pick up the bow and arrow and master the art of archery'),
	('Real World VR', 'real-world-vr', 'Explore the seven wonders of the world in VR');