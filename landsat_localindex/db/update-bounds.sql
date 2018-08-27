--Updates scenes.bounds to the new format that does not use WRS2.

 UPDATE scenes set bounds = q1.safePoly FROM 
(
SELECT product_id as pid, (
case when ((abs(st_x(corner_ul)-st_x(corner_ur)) > 90) OR 
(abs(st_x(corner_ul)-st_x(corner_lr))>90) OR  
(abs(st_x(corner_ul)-st_x(corner_ll))>90)) THEN 
(
st_union(
st_intersection(
st_makeenvelope(-180, -90, 180, 90, 4326),
st_makepolygon(st_makeline(array[st_wrapx(corner_ul, 0, 360), st_wrapx(corner_ur, 0, 360), st_wrapx(corner_lr, 0, 360), st_wrapx(corner_ll, 0, 360), st_wrapx(corner_ul, 0, 360)]))
),
st_intersection(
st_makeenvelope(-180, -90, 180, 90, 4326),
st_makepolygon(st_makeline(array[st_wrapx(corner_ul, 0, -360), st_wrapx(corner_ur, 0, -360), st_wrapx(corner_lr, 0, -360), st_wrapx(corner_ll, 0, -360), st_wrapx(corner_ul, 0, -360)]))
))
) 
else bounds end
)  as safePoly
from scenes) as q1 
WHERE product_id = q1.pid
