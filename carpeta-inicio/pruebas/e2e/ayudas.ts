import { expect, Page } from '@playwright/test';

export const correoAdministrador = process.env.E2E_CORREO ?? 'admin@colegio.pe';
export const claveAdministrador = process.env.E2E_CLAVE ?? 'clave-segura-2026';
export const seccionDemostracion = process.env.E2E_SECCION ?? '3.° A (2026)';

export async function ingresarComoAdministrador(page: Page): Promise<void> {
  await page.goto('/ingreso');
  await page.getByLabel('Correo').fill(correoAdministrador);
  await page.getByLabel('Contraseña').fill(claveAdministrador);
  await page.getByRole('button', { name: 'Ingresar' }).click();
  await expect(page).toHaveURL(/\/alumnos$/);
}

// identificadorUnico evita choques entre ejecuciones repetidas de las pruebas.
export function identificadorUnico(): string {
  const azar = Math.floor(Math.random() * 10_000).toString().padStart(4, '0');
  return '9' + Date.now().toString().slice(-6) + azar;
}

export function horaEnLima(deltaMinutos: number): string {
  const ahora = new Date();
  const partes = new Intl.DateTimeFormat('en-GB', {
    timeZone: 'America/Lima',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).formatToParts(ahora);
  const hora = Number(partes.find((parte) => parte.type === 'hour')?.value ?? '0');
  const minuto = Number(partes.find((parte) => parte.type === 'minute')?.value ?? '0');
  const total = (hora * 60 + minuto + deltaMinutos + 24 * 60) % (24 * 60);
  return `${String(Math.floor(total / 60)).padStart(2, '0')}:${String(total % 60).padStart(2, '0')}`;
}

export function fechaEnLima(): string {
  return new Intl.DateTimeFormat('en-CA', { timeZone: 'America/Lima' }).format(new Date());
}

// fechaEnLimaConDias devuelve la fecha de Lima desplazada en días.
export function fechaEnLimaConDias(dias: number): string {
  const [anio, mes, dia] = fechaEnLima().split('-').map(Number);
  const fecha = new Date(Date.UTC(anio, mes - 1, dia + dias));
  return fecha.toISOString().slice(0, 10);
}

// indiceDiaLima devuelve el día de la semana de Lima: 0 domingo, 6 sábado.
export function indiceDiaLima(): number {
  const [anio, mes, dia] = fechaEnLima().split('-').map(Number);
  return new Date(Date.UTC(anio, mes - 1, dia)).getUTCDay();
}

export async function configurarTurnoManana(
  page: Page,
  ingreso: string,
  limite: string,
  cierre: string,
): Promise<void> {
  await page.goto('/configuracion/turnos');
  const formulario = page.locator('form[action="/configuracion/turnos/manana"]');
  await formulario.getByLabel('Hora de ingreso (Mañana) *').fill(ingreso);
  await formulario.getByLabel('Tolerancia en minutos (Mañana) *').fill('5');
  await formulario.getByLabel('Hora límite de tardanza (Mañana) *').fill(limite);
  await formulario.getByLabel('Hora de cierre (Mañana) *').fill(cierre);
  await formulario.getByLabel('Vigencia desde (Mañana) *').fill(fechaEnLima());
  await formulario.getByRole('button', { name: 'Guardar horario' }).click();
  await expect(page.getByText('El horario del turno se guardó correctamente.')).toBeVisible();
}

export async function crearAlumnoConQR(page: Page): Promise<{ nombre: string; token: string }> {
  const identificador = identificadorUnico();
  await page.goto('/alumnos/nuevo');
  await page.getByLabel('DNI o código *').fill(identificador);
  await page.getByLabel('Nombres *').fill('Luis');
  await page.getByLabel('Apellidos *').fill('Portería ' + identificador.slice(-4));
  await page.getByLabel('Grado y sección *').selectOption({ label: seccionDemostracion });
  await page.getByRole('button', { name: 'Guardar' }).click();
  await expect(page).toHaveURL(/\/alumnos\/\d+\?ok=creado/);

  const generar = page.getByRole('button', { name: 'Generar QR' });
  if (await generar.isVisible()) {
    await generar.click();
    await expect(page.getByText('Código de respaldo')).toBeVisible();
  }
  await page.getByText('Código de respaldo').click();
  const token = (await page.locator('code').first().textContent())?.trim() ?? '';
  return { nombre: `Portería ${identificador.slice(-4)}`, token };
}