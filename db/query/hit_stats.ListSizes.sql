select
	width             as name,
	sum(count) as count
from size_stats
where
	site_id = :site and day >= :start and day <= :end
	{{:filter and path_id in (:filter)}}
group by width
order by count desc, name asc
