-- Skrip reset untuk uji beban
DELETE FROM enrollments
WHERE user_id IN (SELECT id FROM users WHERE username LIKE 'mhs%');

UPDATE courses c
SET enrolled_count = (SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id);

-- Reset bimbingan tugas akhir (Supervisor)
DELETE FROM supervisions;
UPDATE supervisors SET supervised_count = 0;

