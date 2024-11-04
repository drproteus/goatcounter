select
	trim(name || ' ' || version) as name,
	sum(count)            as count
from browser_stats
join browsers using (browser_id)
where
	site_id = :site and day >= :start and day <= :end and
	{{:filter path_id in (:filter) and}}
	lower(name) = lower(:browser)
group by name, version
order by count desc, name asc
limit :limit offset :offset
